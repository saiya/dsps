package outgoing

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockResponse struct {
	http.Response
	body *mockResponseBody
}

func newMockResponse(status int, body []byte) *mockResponse {
	rb := &mockResponseBody{r: bytes.NewReader(body)}
	return &mockResponse{
		Response: http.Response{
			StatusCode: status,
			Body:       rb,
		},
		body: rb,
	}
}

func (res *mockResponse) assertProperlyClosed(t *testing.T) {
	if res.body.ErrOnRead == nil {
		_, err := res.body.r.Read(make([]byte, 1))
		assert.Equal(t, io.EOF, err, "response body must be read until EOF")
	}
	assert.Equal(t, true, res.body.isClosed, "response body should be closed")
}

type mockResponseBody struct {
	isClosed bool

	ErrOnRead  error
	ErrOnClose error

	r io.Reader
}

func (body *mockResponseBody) Read(p []byte) (n int, err error) {
	if body.ErrOnRead != nil {
		return 0, body.ErrOnRead
	}
	return body.r.Read(p)
}

func (body *mockResponseBody) Close() error {
	if body.isClosed {
		panic("mockResponseBody closed twice (io.Closer forbids this)")
	}
	body.isClosed = true
	return body.ErrOnClose
}
