package middleware_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/http"
	. "github.com/saiya/dsps/server/http/testing"
)

func TestAdminAuthMiddleware(t *testing.T) {
	WithServer(t, `
admin:
	auth:
		bearer:
			- 'my-api-key'
	`, func(deps *ServerDependencies) {}, func(deps *ServerDependencies, baseURL string) {
		assert.Equal(t, `Bearer my-api-key`, AdminAuthHeaders(t, deps)["Authorization"])

		// Pass
		res := DoHTTPRequestWithHeaders(t, "PUT", fmt.Sprintf("%s/admin/log/level?category=auth&level=ERROR", baseURL), AdminAuthHeaders(t, deps), ``)
		assert.NoError(t, res.Body.Close())
		assert.Equal(t, 204, res.StatusCode)

		// No API Key
		res = DoHTTPRequestWithHeaders(t, "PUT", fmt.Sprintf("%s/admin/log/level?category=auth&level=ERROR", baseURL), map[string]string{}, ``)
		assert.NoError(t, res.Body.Close())
		assert.Equal(t, 403, res.StatusCode)
	})

	WithServer(t, `
admin:
	auth:
		bearer:
			- 'my-api-key'
		networks:
			- 0.0.0.0/32
	`, func(deps *ServerDependencies) {}, func(deps *ServerDependencies, baseURL string) {
		assert.Equal(t, `Bearer my-api-key`, AdminAuthHeaders(t, deps)["Authorization"])

		// IP range not trusted
		res := DoHTTPRequestWithHeaders(t, "PUT", fmt.Sprintf("%s/admin/log/level?category=auth&level=ERROR", baseURL), AdminAuthHeaders(t, deps), ``)
		assert.NoError(t, res.Body.Close())
		assert.Equal(t, 403, res.StatusCode)
	})
}
