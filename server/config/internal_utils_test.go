package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeDuration(t *testing.T) {
	assert.Panics(t, func() {
		makeDuration("INVALID")
	})
}
