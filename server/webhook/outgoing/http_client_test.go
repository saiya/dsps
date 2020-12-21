package outgoing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	"github.com/stretchr/testify/assert"
)

func TestHTTPClient(t *testing.T) {
	maxConns := 1234
	idleConnTimeout := 123 * time.Second
	maxRedirects := 0
	c := newHTTPClientFor(context.Background(), &config.OutgoingWebhookConfig{
		Connection: config.OutgoingWebhookConnectionConfig{
			Max:         &maxConns,
			MaxIdleTime: &domain.Duration{Duration: idleConnTimeout},
		},
		MaxRedirects: &maxRedirects,
	})

	tr := c.Transport.(*http.Transport)
	assert.Equal(t, maxConns, tr.MaxIdleConns)
	assert.Equal(t, maxConns, tr.MaxIdleConnsPerHost)
	assert.Equal(t, maxConns, tr.MaxConnsPerHost)
	assert.Equal(t, idleConnTimeout, tr.IdleConnTimeout)
}

func TestHTTPClientRedirect(t *testing.T) {
	ctx := context.Background()

	called := 0
	redirectUntil := 0
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		called++
		if called < redirectUntil {
			rw.Header().Set("Location", serverURL)
			rw.WriteHeader(307)
		} else {
			rw.WriteHeader(204)
		}
	}))
	serverURL = server.URL
	defer server.Close()

	maxConns := 1
	idleConnTimeout := 5 * time.Second
	maxRedirects := 3
	c := newHTTPClientFor(context.Background(), &config.OutgoingWebhookConfig{
		Connection: config.OutgoingWebhookConnectionConfig{
			Max:         &maxConns,
			MaxIdleTime: &domain.Duration{Duration: idleConnTimeout},
		},
		MaxRedirects: &maxRedirects,
	})

	// Without redirect
	called = 0
	redirectUntil = 0
	req, err := http.NewRequestWithContext(ctx, "GET", serverURL, nil)
	assert.NoError(t, err)
	res, err := c.Do(req)
	assert.Equal(t, 1, called)
	assert.NoError(t, res.Body.Close())
	assert.NoError(t, err)
	assert.Equal(t, 204, res.StatusCode)

	// With redirect (<= maxRedirects)
	called = 0
	redirectUntil = maxRedirects
	req, err = http.NewRequestWithContext(ctx, "GET", serverURL, nil)
	assert.NoError(t, err)
	res, err = c.Do(req)
	assert.Equal(t, maxRedirects, called)
	assert.NoError(t, res.Body.Close())
	assert.NoError(t, err)
	assert.Equal(t, 204, res.StatusCode)

	// With redirect (> maxRedirects)
	called = 0
	redirectUntil = maxRedirects + 1
	req, err = http.NewRequestWithContext(ctx, "GET", serverURL, nil)
	assert.NoError(t, err)
	res, err = c.Do(req)
	assert.Equal(t, maxRedirects, called)
	assert.NoError(t, res.Body.Close())
	assert.Regexp(t, `too many redirects`, err.Error())
	assert.Equal(t, 307, res.StatusCode)
}
