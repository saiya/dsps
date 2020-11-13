package internal_test

import (
	"testing"

	"github.com/saiya/dsps/server/domain"
	. "github.com/saiya/dsps/server/storage/internal"
	. "github.com/saiya/dsps/server/testing"
)

func TestMalformedHandle(t *testing.T) {
	// TODO: Test DecodeAckHandle also

	_, err := DecodeMultiplexAckHandle(domain.AckHandle{
		SubscriberLocator: domain.SubscriberLocator{
			ChannelID:    "ch1",
			SubscriberID: "s1",
		},
		Handle: "xxx",
	})
	IsError(t, domain.ErrMalformedAckHandle, err)
}
