package testing

import (
	"context"
	"fmt"

	"github.com/saiya/dsps/server/domain"
	dspstesting "github.com/saiya/dsps/server/testing"
)

// DisabledChannelID is ChannelID that StubChannelProvider always rejects.
var DisabledChannelID domain.ChannelID = "disabled-channel"

// StubChannelExpire is expire (TTL) of any channels pro
var StubChannelExpire = dspstesting.MakeDuration("5m")

// StubChannelProvider is simple stub implementation of ChannelProvider
var StubChannelProvider domain.ChannelProvider = dspstesting.ChannelProviderFunc(func(id domain.ChannelID) (domain.Channel, error) {
	if id == DisabledChannelID {
		return nil, domain.ErrInvalidChannel
	}
	return &stubChannel{
		id:     id,
		expire: StubChannelExpire,
	}, nil
})

type stubChannel struct {
	id     domain.ChannelID
	expire domain.Duration
}

func (c *stubChannel) String() string {
	return fmt.Sprintf("StubChannel(%s)", c.id)
}

func (c *stubChannel) Expire() domain.Duration {
	return c.expire
}

func (c *stubChannel) ValidateJwt(ctx context.Context, jwt string) error {
	return nil
}

func (c *stubChannel) SendOutgoingWebhook(ctx context.Context, msg domain.Message) error {
	return nil
}
