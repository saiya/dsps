package domain

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// SubscriberID is ID of the subscriber, unique within channel
type SubscriberID string

// SubscriberLocator is unique identifier of the PubSub subscriber, unique within channel
type SubscriberLocator struct {
	ChannelID    ChannelID
	SubscriberID SubscriberID
}

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

// AckHandle is an token to remove received (acknowledged) messages from a subscriber.
type AckHandle struct {
	SubscriberLocator
	Handle string
}

// see: doc/interface/validation_rule.md
var (
	subscriberIDRegexp = regexp.MustCompile("^[0-9a-z][0-9a-z_-]{0,62}$")
	messageIDRegexp    = regexp.MustCompile("^[0-9a-z][0-9a-z_-]{0,62}$")
)

// ParseSubscriberID try to parse ID
func ParseSubscriberID(str string) (SubscriberID, error) {
	if !subscriberIDRegexp.MatchString(str) {
		return SubscriberID(""), fmt.Errorf("SubscriberID must match with %s", subscriberIDRegexp.String())
	}
	return SubscriberID(str), nil
}

// ParseMessageID try to parse ID
func ParseMessageID(str string) (MessageID, error) {
	if !messageIDRegexp.MatchString(str) {
		return MessageID(""), fmt.Errorf("MessageID must match with %s", messageIDRegexp.String())
	}
	return MessageID(str), nil
}
