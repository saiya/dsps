package testing

import (
	"github.com/saiya/dsps/server/domain"
	dspstesting "github.com/saiya/dsps/server/testing"
)

// DisabledChannelID is ChannelID that StubChannelProvider always rejects.
var DisabledChannelID domain.ChannelID = "disabled-channel"

// StubChannelProvider is simple stub implementation of ChannelProvider
var StubChannelProvider domain.ChannelProvider = func(id domain.ChannelID) domain.Channel {
	if id == DisabledChannelID {
		return nil
	}
	return &stubChannel{
		expire: dspstesting.MakeDuration("5m"),
	}
}

type stubChannel struct {
	expire domain.Duration
}

func (c *stubChannel) Expire() domain.Duration {
	return c.expire
}
