package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/domain"
)

// MessagesEqual compares list of Messages by it's content
func MessagesEqual(t *testing.T, expected []domain.Message, actual []domain.Message) {
	assert.EqualValues(t, expected, actual)
}
