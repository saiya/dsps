package domain

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// MessageID is ID of the message, unique within channel
type MessageID string

// MessageLocator is unique identifier of the Message, unique within channel
type MessageLocator struct {
	ChannelID ChannelID
	MessageID MessageID
}

// Message is an atomic datagram of the PubSub communication
type Message struct {
	MessageLocator
	Content json.RawMessage
}

// see: doc/interface/validation_rule.md
var messageIDRegexp = regexp.MustCompile("^[0-9a-z][0-9a-z_-]{0,62}$")

// ParseMessageID try to parse ID
func ParseMessageID(str string) (MessageID, error) {
	if !messageIDRegexp.MatchString(str) {
		return MessageID(""), fmt.Errorf("MessageID must match with %s", messageIDRegexp.String())
	}
	return MessageID(str), nil
}

// BelongsToSameChannel returns false if messages belongs to various channels
func BelongsToSameChannel(msgs []Message) bool {
	if len(msgs) == 0 {
		return true
	}

	ch := msgs[0].ChannelID
	for _, msg := range msgs {
		if msg.ChannelID != ch {
			return false
		}
	}
	return true
}
