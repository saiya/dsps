package outgoing

import (
	"context"
	"net/http"

	"github.com/saiya/dsps/server/config"
	"golang.org/x/xerrors"
)

func newHTTPClientFor(ctx context.Context, cfg *config.OutgoingWebhookConfig) *http.Client {
	tr := &http.Transport{
		MaxIdleConns:        *cfg.Connection.Max,
		MaxIdleConnsPerHost: *cfg.Connection.Max,
		MaxConnsPerHost:     *cfg.Connection.Max,

		IdleConnTimeout: cfg.Connection.MaxIdleTime.Duration,
	}
	maxRedirects := *cfg.MaxRedirects
	return &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return xerrors.Errorf("too many redirects")
			}
			return nil
		},
	}
}
