package outgoing

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/logger"
	"github.com/saiya/dsps/server/sentry"
	"github.com/saiya/dsps/server/telemetry"
)

// Client is an outgoing-webhook client.
type Client interface {
	Send(ctx context.Context, msg domain.Message) error

	// Shutdown this client.
	// This method wait until all in-flight request ends.
	Close(ctx context.Context)

	String() string
}

type clientImpl struct {
	_isClosed int32 // 0: available, 1: closing

	method  string
	url     string
	headers map[string]string

	timeout time.Duration
	retry   retry

	h         *http.Client // Note that Client does not own this object, ClientTemplate owns.
	telemetry *telemetry.Telemetry
	sentry    sentry.Sentry
}

func newClientImpl(tpl *clientTemplate, tplEnv domain.TemplateStringEnv) (Client, error) {
	c := &clientImpl{
		_isClosed: 0,

		method:  tpl.Method,
		headers: make(map[string]string, len(tpl.Headers)),

		timeout: tpl.Timeout.Duration,
		retry:   newRetry(&tpl.Retry),

		h:         tpl.h,
		telemetry: tpl.telemetry,
		sentry:    tpl.sentry,
	}

	var err error
	c.url, err = tpl.URL.Execute(tplEnv)
	if err != nil {
		return nil, xerrors.Errorf(`failed to expand template of webhook URL "%s": %w`, tpl.URL, err)
	}
	for name, valueTpl := range tpl.Headers {
		c.headers[name], err = valueTpl.Execute(tplEnv)
		if err != nil {
			return nil, xerrors.Errorf(`failed to expand template of webhook header "%s", "%s": %w`, name, valueTpl, err)
		}
	}
	return c, nil
}

func (c *clientImpl) String() string {
	return fmt.Sprintf("%s %s", c.method, c.url)
}

func (c *clientImpl) Send(ctx context.Context, msg domain.Message) error {
	if c.isClosed() {
		return xerrors.Errorf("outgoing-webhook client already closed")
	}
	logger.Of(ctx).Debugf(logger.CatOutgoingWebhook, "sending outgoing webhook (channel: %s, messageID: %s) to %s", msg.ChannelID, msg.MessageID, c.url)

	body, err := encodeWebhookBody(ctx, msg)
	if err != nil {
		return xerrors.Errorf("failed to generate outgoing webhook body: %w", err)
	}

	return c.retry.Do(ctx, c.sentry, fmt.Sprintf("outgoing-webhook to %s", c.url), func() (*http.Request, *http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, c.method, c.url, strings.NewReader(body))
		if err != nil {
			return req, nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		for name, value := range c.headers {
			// Should overwrite default headers, thus use Set() rather than Add()
			req.Header.Set(name, value)
		}

		ctx, end := c.telemetry.StartHTTPSpan(ctx, false, req)
		defer end()
		res, err := c.h.Do(req)
		if res != nil {
			logger.Of(ctx).Debugf(logger.CatOutgoingWebhook, "received outgoing webhook response (%s %d, contentLength: %d)", res.Proto, res.StatusCode, res.ContentLength)
			c.telemetry.SetHTTPResponseAttributes(ctx, res.StatusCode, res.ContentLength)
		}
		return req, res, err
	})
}

func (c *clientImpl) Close(ctx context.Context) {
	if atomic.CompareAndSwapInt32(&c._isClosed, 0, 1) {
		return
	}
}

func (c *clientImpl) isClosed() bool {
	return atomic.LoadInt32(&c._isClosed) != 0
}
