package testing

import "github.com/saiya/dsps/server/domain"

// ChannelProviderFunc wraps function as ChannelProvider
type ChannelProviderFunc func(id domain.ChannelID) (domain.Channel, error)

// Get implements ChannelProvider
func (f ChannelProviderFunc) Get(id domain.ChannelID) (domain.Channel, error) {
	return f(id)
}

// GetFileDescriptorPressure implements ChannelProvider
func (f ChannelProviderFunc) GetFileDescriptorPressure() int {
	return 0
}
