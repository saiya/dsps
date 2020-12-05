package redis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCorruptedMessageEnvelope(t *testing.T) {
	raw := `INVALID JSON`
	_, err := unwrapMessage("ch-1", raw)
	assert.Contains(t, err.Error(), "Failed to parse message envelope JSON")
}
