package domain

import (
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

// see: doc/interface/validation_rule.md
var subscriberIDRegexp = regexp.MustCompile("^[0-9a-z][0-9a-z_-]{0,62}$")

// ParseSubscriberID try to parse ID
func ParseSubscriberID(str string) (SubscriberID, error) {
	if !subscriberIDRegexp.MatchString(str) {
		return SubscriberID(""), fmt.Errorf("SubscriberID must match with %s", subscriberIDRegexp.String())
	}
	return SubscriberID(str), nil
}
