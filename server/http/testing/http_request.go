package testing

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	dspshttp "github.com/saiya/dsps/server/http"
)

// DoHTTPRequest performs HTTP request
func DoHTTPRequest(t *testing.T, method string, url string, body string) *http.Response {
	return DoHTTPRequestWithHeaders(t, method, url, make(map[string]string), body)
}

// DoHTTPRequestWithHeaders performs HTTP request
func DoHTTPRequestWithHeaders(t *testing.T, method string, url string, headers map[string]string, body string) *http.Response {
	var bodyR io.Reader
	if body != "" {
		bodyR = strings.NewReader(body)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, url, bodyR)
	assert.NoError(t, err)
	for name, value := range headers {
		req.Header.Add(name, value)
	}

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	return res
}

// AdminAuthHeaders make HTTP headers for Admin API call
func AdminAuthHeaders(t *testing.T, deps *dspshttp.ServerDependencies) map[string]string {
	return map[string]string{"Authorization": "Bearer " + deps.Config.Admin.Auth.BearerTokens[0]}
}
