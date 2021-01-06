package testing

import (
	"context"

	"github.com/saiya/dsps/server/domain"
)

// ChannelProviderFunc wraps function as ChannelProvider
type ChannelProviderFunc func(id domain.ChannelID) (domain.Channel, error)

// Shutdown closes ChannelProvider
func (f ChannelProviderFunc) Shutdown(ctx context.Context) {}

// Get implements ChannelProvider
func (f ChannelProviderFunc) Get(id domain.ChannelID) (domain.Channel, error) {
	return f(id)
}

// GetFileDescriptorPressure implements ChannelProvider
func (f ChannelProviderFunc) GetFileDescriptorPressure() int {
	return 0
}

// JWTClockSkewLeewayMax implements ChannelProvider
func (f ChannelProviderFunc) JWTClockSkewLeewayMax() domain.Duration {
	return domain.Duration{}
}
