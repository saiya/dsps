package domain_test

import (
	"errors"
	"testing"

	. "github.com/saiya/dsps/server/domain"
	"github.com/stretchr/testify/assert"
)

func TestIsStorageNonFatalError(t *testing.T) {
	for _, err := range []error{ErrInvalidChannel, ErrSubscriptionNotFound, ErrMalformedAckHandle} {
		assert.True(t, IsStorageNonFatalError(err))
	}
	assert.False(t, IsStorageNonFatalError(errors.New(`test error`)))
}
