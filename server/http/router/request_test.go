package router_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/http/router"
)

func TestRequest(t *testing.T) {
	body, err := Request{Request: httptest.NewRequest("GET", "/", strings.NewReader(`{"hi":"hello"}`))}.ReadBody()
	assert.NoError(t, err)
	assert.Equal(t, `{"hi":"hello"}`, string(body))

	assert.Equal(t, "bar baz", Request{Request: httptest.NewRequest("GET", "/?foo=bar%20baz", strings.NewReader(``))}.GetQueryParam("foo"))
	assert.Equal(t, "", Request{Request: httptest.NewRequest("GET", "/?foo=bar", strings.NewReader(``))}.GetQueryParam("baz"))
}
