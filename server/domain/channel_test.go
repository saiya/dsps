package domain_test

import (
	"testing"

	. "github.com/saiya/dsps/server/domain"
	"github.com/stretchr/testify/assert"
)

func TestParseChannelID(t *testing.T) {
	errorMsg := `ChannelID must match with ^[0-9a-z][0-9a-z_-]{0,62}$`

	id, err := ParseChannelID(`my-channel-123`)
	assert.NoError(t, err)
	assert.Equal(t, `my-channel-123`, string(id))

	_, err = ParseChannelID(``)
	assert.Errorf(t, err, errorMsg)

	_, err = ParseChannelID(`INVALID`)
	assert.Errorf(t, err, errorMsg)
}
