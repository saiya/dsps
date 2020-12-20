package domain

import (
	"context"
	"fmt"
	"regexp"
)

// ChannelID is ID of the PubSub channel, system-wide unique value
type ChannelID string

//go:generate mockgen -source=${GOFILE} -package=mock -destination=./mock/${GOFILE}

// ChannelProvider provides configured Channel object.
// If given ChannelID is not valid for this server process, returns (nil, domain.ErrInvalidChannel).
type ChannelProvider interface {
	Get(id ChannelID) (Channel, error)

	GetFileDescriptorPressure() int
}

// Channel struct holds all objects/information of a channel
type Channel interface {
	Expire() Duration

	// Note that this method does not check revocation list.
	ValidateJwt(ctx context.Context, jwt string) error

	SendOutgoingWebhook(ctx context.Context, msg Message) error
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
