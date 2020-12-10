package testing

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// IsError asserts error object
func IsError(t *testing.T, expected error, actual error) bool {
	if errors.Is(actual, expected) {
		return true
	}
	assert.Fail(t, "error unmatch", "expected %#v but %#v", expected, actual)
	return false
}

// IsOneOfErrors asserts error object
func IsOneOfErrors(t *testing.T, expected []error, actual error) bool {
	for _, allowed := range expected {
		if errors.Is(actual, allowed) {
			return true
		}
	}
	assert.Fail(t, "error unmatch", "expected %#v but %#v", expected, actual)
	return false
}
