package testing

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// DoHTTPRequest perform HTTP request
func DoHTTPRequest(t *testing.T, method string, url string, body string) *http.Response {
	var bodyR io.Reader
	if body != "" {
		bodyR = strings.NewReader(body)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, url, bodyR)
	assert.NoError(t, err)
	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	return res
}
