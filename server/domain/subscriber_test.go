package domain_test

import (
	"testing"

	. "github.com/saiya/dsps/server/domain"
	"github.com/stretchr/testify/assert"
)

func TestParseSubscriberID(t *testing.T) {
	errorMsg := `SubscriberID must match with ^[0-9a-z][0-9a-z_-]{0,62}$`

	id, err := ParseSubscriberID(`my-sbsc-123`)
	assert.NoError(t, err)
	assert.Equal(t, `my-sbsc-123`, string(id))

	_, err = ParseSubscriberID(``)
	assert.Errorf(t, err, errorMsg)

	_, err = ParseSubscriberID(`INVALID`)
	assert.Errorf(t, err, errorMsg)
}
