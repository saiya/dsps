package middleware_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"

	. "github.com/saiya/dsps/server/http"
	. "github.com/saiya/dsps/server/http/testing"
	"github.com/saiya/dsps/server/telemetry"
)

func TestTracingMiddleware(t *testing.T) {
	var host string
	var responseLength int64
	tr := telemetry.WithStubTracing(t, func(telemetry *telemetry.Telemetry) {
		WithServer(t, `logging: { category: { "*": ERROR, http: INFO } }`, func(deps *ServerDependencies) {
			deps.Telemetry = telemetry
		}, func(deps *ServerDependencies, baseURL string) {
			baseURLp, err := url.Parse(baseURL)
			assert.NoError(t, err)
			host = baseURLp.Host

			url := baseURL + "/probe/liveness?param=value"
			res := DoHTTPRequestWithHeaders(t, "GET", url, map[string]string{
				"User-Agent": "tracing-test/1.0",
			}, ``)
			assert.NoError(t, res.Body.Close())
			assert.Equal(t, 200, res.StatusCode)
			responseLength = res.ContentLength
		})
	})
	tr.OT.AssertSpan(0, trace.SpanKindServer, "HTTP GET /probe/liveness", map[string]interface{}{
		"http.method":                  "GET",
		"http.scheme":                  "http",
		"http.host":                    host,
		"http.target":                  "/probe/liveness?param=value",
		"http.user_agent":              "tracing-test/1.0",
		"http.request_content_length":  int64(0),
		"http.route":                   "/probe/liveness",
		"http.client_ip":               "127.0.0.1",
		"http.status_code":             int64(200),
		"http.response_content_length": responseLength,
	})
}
