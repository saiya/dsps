package middleware_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/http"
	. "github.com/saiya/dsps/server/http/testing"
)

func TestDefaultHeadersFilter(t *testing.T) {
	WithServer(t, `
http:
	defaultHeaders:
		Strict-Transport-Security: "max-age=31536000 ; includeSubDomains"
		X-Frame-Options: ""
	`, func(deps *ServerDependencies) {}, func(deps *ServerDependencies, baseURL string) {
		res := DoHTTPRequest(t, "GET", baseURL+"/probe/liveness", ``)
		assert.NoError(t, res.Body.Close())
		assert.Equal(t, 200, res.StatusCode)

		// Enabled by default
		assert.Equal(t, `nosniff`, res.Header.Get("X-Content-Type-Options"))

		// Changed by user config
		assert.Equal(t, `max-age=31536000 ; includeSubDomains`, res.Header.Get("Strict-Transport-Security"))
		assert.Equal(t, 0, len(res.Header.Values("X-Frame-Options")))
	})
}
