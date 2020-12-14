package router_test

import (
	"net/http/httptest"
	"testing"

	. "github.com/saiya/dsps/server/http/router"
	"github.com/stretchr/testify/assert"
)

func TestResponseWriter(t *testing.T) {
	w := NewResponseWriter(httptest.NewRecorder())
	assert.Equal(t, ResponseWritten{StatusCode: 200, BodyBytes: 0}, w.Written())

	w = NewResponseWriter(httptest.NewRecorder())
	w.WriteHeader(400)
	_, err := w.Write([]byte{1, 2, 3})
	assert.NoError(t, err)
	_, err = w.Write([]byte{4, 5})
	assert.NoError(t, err)
	assert.Equal(t, ResponseWritten{StatusCode: 400, BodyBytes: 5}, w.Written())
}
