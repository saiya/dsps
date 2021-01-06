package redis

import (
	"fmt"

	"github.com/saiya/dsps/server/domain"
)

type channelKeys struct {
	// All keys must be prefixed with {channel-id} due to partioning.
	channelID domain.ChannelID
}

func keyOfChannel(channelID domain.ChannelID) channelKeys {
	return channelKeys{channelID: channelID}
}

// type of value is channelClock
func (rk channelKeys) Clock() string {
	return fmt.Sprintf("c.{%s}.clock", rk.channelID)
}

// type of value is channelClock
func (rk channelKeys) SubscriberCursor(rcv domain.SubscriberID) string {
	return fmt.Sprintf("c.{%s}.r.%s", rk.channelID, rcv)
}

// type of value is JSON
func (rk channelKeys) MessageBodyPrefix() string {
	return fmt.Sprintf("c.{%s}.m.", rk.channelID)
}

// type of value is JSON
func (rk channelKeys) MessageBody(clock channelClock) string {
	// MUST start with MessageBodyPrefix()
	return fmt.Sprintf("c.{%s}.m.%d", rk.channelID, clock)
}

// type of value is channelClock
func (rk channelKeys) MessageDedup(id domain.MessageID) string {
	return fmt.Sprintf("c.{%s}.mid.%s", rk.channelID, id)
}

type jtiKeys struct {
	jti domain.JwtJti
}

func keyOfJti(jti domain.JwtJti) jtiKeys {
	return jtiKeys{jti: jti}
}

func (jti jtiKeys) Revocation() string {
	return fmt.Sprintf("jwt.{%s}.revoke", jti.jti)
}
