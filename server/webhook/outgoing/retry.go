package outgoing

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"time"

	sentrygo "github.com/getsentry/sentry-go"
	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/logger"
	"github.com/saiya/dsps/server/sentry"
)

type retry struct {
	count              int
	interval           time.Duration
	intervalMultiplier float64
	intervalJitter     time.Duration

	jitterLock sync.Mutex
	jitter     *rand.Rand
}

func newRetry(cfg *config.OutgoingWebhookRetryConfig) retry {
	return retry{
		count:              *cfg.Count,
		interval:           cfg.Interval.Duration,
		intervalMultiplier: *cfg.IntervalMultiplier,
		intervalJitter:     cfg.IntervalJitter.Duration,
	}
}

// Do wraps given operation with retry handling.
// Callback function should not close response body stream, this method closes it.
func (r *retry) Do(ctx context.Context, sentryInstance sentry.Sentry, description string, f func() (*http.Request, *http.Response, error)) error {
	attempt := 0
	for {
		req, res, err := f()
		{
			// for "http" type, see https://github.com/getsentry/sentry-docs/issues/1709#issue-624545593
			sentryData := make(map[string]interface{})
			if req != nil {
				sentryData["method"] = req.Method
			}
			if res != nil {
				if res.Request != nil {
					sentryData["url"] = res.Request.URL
				}
				sentryData["status_code"] = res.StatusCode
			}
			sentry.AddBreadcrumb(ctx, &sentrygo.Breadcrumb{
				Type:    "http",
				Level:   sentrygo.LevelInfo,
				Message: "Outgoing webhook",
				Data:    sentryData,
			})
		}
		if res != nil {
			// Should read all response body otherwise disrupts keep-alive.
			if _, copyErr := io.Copy(ioutil.Discard, res.Body); copyErr != nil {
				logger.Of(ctx).Debugf(logger.CatOutgoingWebhook, "failed to read response body: %w", copyErr)
			}
			if closeErr := res.Body.Close(); closeErr != nil {
				logger.Of(ctx).Debugf(logger.CatOutgoingWebhook, "failed to close response body stream: %w", closeErr)
			}

			if err == nil && 200 <= res.StatusCode && res.StatusCode <= 299 {
				return nil // Success
			}
		}
		attempt++ // Failed

		var shouldRetry bool
		shouldRetry, err = r.postprocess(res, err) // always returns non-nil error object
		if attempt > r.count || !shouldRetry {
			logger.Of(ctx).Warnf(logger.CatOutgoingWebhook, "outgoing webhook failed: %w", err)
			sentry.RecordError(ctx, fmt.Errorf("outgoing webhook failed: %w", err))
			return err
		}

		wait := r.computeRetryWait(attempt)
		logger.Of(ctx).Infof(logger.CatOutgoingWebhook, "retrying outgoing webhook after %s: %w", wait, err)
		time.Sleep(wait)
		continue
	}
}

func (r *retry) computeRetryWait(attempt int) time.Duration {
	ns := float64(r.interval.Nanoseconds()) * math.Pow(r.intervalMultiplier, float64(attempt))

	r.jitterLock.Lock()
	defer r.jitterLock.Unlock()
	if r.jitter == nil {
		r.jitter = rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	}
	ns += (r.jitter.Float64()*2 - 1) * float64(r.intervalJitter.Nanoseconds())

	return time.Duration(math.Round(ns)) * time.Nanosecond
}

// See server/doc/outgoing-webhook.md for spec.
var statusCodeToRetry = map[int]bool{
	// 400 to 418
	400: true,  // Bad Request
	401: false, // Unauthorized
	402: false, // Payment Required
	403: false, // Forbidden
	404: true,  // Not Found
	405: false, // Method Not Allowed
	406: false, // Not Acceptable
	407: false, // Proxy Authentication Required
	408: true,  // Request Timeout
	409: true,  // Conflict
	410: false, // Gone
	411: false, // Length Required
	412: false, // Precondition Failed
	413: false, // Payload Too Large
	414: false, // URI Too Long
	415: false, // Unsupported Media Type
	416: false, // Range Not Satisfiable
	417: false, // Expectation Failed
	418: false, // I'm a teapot

	// Other 4xx
	426: false, // Upgrade Required
	431: false, // Request Header Fields Too Large
	451: false, // Unavailable For Legal Reasons

	// 5xx
	501: false, // Not Implemented
}

// always returns non-nil error object describes failure.
func (r *retry) postprocess(res *http.Response, err error) (shouldRetry bool, errWrapped error) {
	if err != nil {
		return true, err
	}
	if res == nil {
		return true, xerrors.Errorf("no response object returned")
	}

	if result, matched := statusCodeToRetry[res.StatusCode]; matched {
		return result, xerrors.Errorf("status code %d returned", res.StatusCode)
	}
	return true, xerrors.Errorf("status code %d returned", res.StatusCode)
}
