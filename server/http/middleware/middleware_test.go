package middleware_test

import (
	"context"
	"testing"

	"github.com/saiya/dsps/server/http/router"
	"github.com/stretchr/testify/assert"
)

func withNextFunc(t *testing.T, expectCalled bool, f func(next func(context.Context, router.MiddlewareArgs))) {
	called := false
	f(func(context.Context, router.MiddlewareArgs) {
		called = true
	})
	assert.Equal(t, expectCalled, called)
}
