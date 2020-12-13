package domain_test

import (
	"encoding/json"
	"testing"

	. "github.com/saiya/dsps/server/domain"
	"github.com/stretchr/testify/assert"
)

func TestParseMessageID(t *testing.T) {
	errorMsg := `MessageID must match with ^[0-9a-z][0-9a-z_-]{0,62}$`

	id, err := ParseMessageID(`my-channel-123`)
	assert.NoError(t, err)
	assert.Equal(t, `my-channel-123`, string(id))

	_, err = ParseMessageID(``)
	assert.Errorf(t, err, errorMsg)

	_, err = ParseMessageID(`INVALID`)
	assert.Errorf(t, err, errorMsg)
}

func TestBelongsToSameChannel(t *testing.T) {
	assert.True(t, BelongsToSameChannel([]Message{}))
	assert.True(t, BelongsToSameChannel([]Message{
		{MessageLocator: MessageLocator{ChannelID: "ch-1", MessageID: "msg-1"}, Content: json.RawMessage(`{}`)},
		{MessageLocator: MessageLocator{ChannelID: "ch-1", MessageID: "msg-2"}, Content: json.RawMessage(`{}`)},
	}))
	assert.False(t, BelongsToSameChannel([]Message{
		{MessageLocator: MessageLocator{ChannelID: "ch-1", MessageID: "msg-1"}, Content: json.RawMessage(`{}`)},
		{MessageLocator: MessageLocator{ChannelID: "ch-xxx", MessageID: "msg-2"}, Content: json.RawMessage(`{}`)},
	}))
}
