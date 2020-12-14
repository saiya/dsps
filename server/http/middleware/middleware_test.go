package middleware_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func withNextFunc(t *testing.T, expectCalled bool, f func(next func(context.Context))) {
	called := false
	f(func(context.Context) {
		called = true
	})
	assert.Equal(t, expectCalled, called)
}
