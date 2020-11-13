package domain

import (
	"fmt"
	"regexp"
)

// ChannelID is ID of the PubSub channel, system-wide unique value
type ChannelID string

// ChannelProvider provides configured Channel object.
// If given ChannelID is not valid for this server process, returns nil.
type ChannelProvider func(id ChannelID) *Channel

// Channel struct holds all objects/information of a channel
type Channel struct {
	Expire Duration
}

// see: doc/interface/validation_rule.md
var (
	channelIDRegexp = regexp.MustCompile("^[0-9a-z][0-9a-z_-]{0,62}$")
)

// ParseChannelID try to parse ID
func ParseChannelID(str string) (ChannelID, error) {
	if !channelIDRegexp.MatchString(str) {
		return ChannelID(""), fmt.Errorf("ChannelID must match with %s", channelIDRegexp.String())
	}
	return ChannelID(str), nil
}
