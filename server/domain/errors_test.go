package domain_test

import (
	"errors"
	"testing"

	. "github.com/saiya/dsps/server/domain"
	"github.com/stretchr/testify/assert"
)

func TestErrorWithCode(t *testing.T) {
	err := NewErrorWithCode("test.error.code")
	assert.Equal(t, "test.error.code", err.Code())
	assert.Equal(t, "test.error.code", err.Error())

	assert.Nil(t, err.Unwrap())
}

func TestWrapErrorWithCode(t *testing.T) {
	wrapped := errors.New(`wrapped error`)
	err := WrapErrorWithCode("test.error.code", wrapped)

	assert.Equal(t, "test.error.code", err.Code())
	assert.Equal(t, `wrapped error`, err.Error())

	assert.Same(t, wrapped, err.Unwrap())
	assert.True(t, errors.Is(err, wrapped))
}
