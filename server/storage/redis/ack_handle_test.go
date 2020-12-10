package redis

import (
	"encoding/json"
	"testing"

	"github.com/saiya/dsps/server/domain"
	dspstesting "github.com/saiya/dsps/server/testing"
	"github.com/stretchr/testify/assert"
)

func TestAckHandleChecksum(t *testing.T) {
	sl := domain.SubscriberLocator{
		ChannelID:    "ch-1",
		SubscriberID: "sbsc-1",
	}
	for _, data := range []ackHandleData{
		{LastMessageClock: 0},
		{LastMessageClock: 1024},
		{LastMessageClock: -1024},
		{LastMessageClock: clockMin},
		{LastMessageClock: clockMax},
	} {
		data.Checksum = "INVALID"
		data.ComputeChechsum(sl)
		assert.Equal(t, "INVALID", data.Checksum, "ComputeChechsum() should not modify struct")

		assert.Equal(t, data.ComputeChechsum(sl), data.ComputeChechsum(sl))
		assert.NotEqual(t, data.ComputeChechsum(sl), ackHandleData{LastMessageClock: data.LastMessageClock - 1}.ComputeChechsum(sl))
		assert.NotEqual(t, data.ComputeChechsum(sl), ackHandleData{LastMessageClock: data.LastMessageClock + 1}.ComputeChechsum(sl))
		assert.NotEqual(t, data.ComputeChechsum(sl), data.ComputeChechsum(domain.SubscriberLocator{
			ChannelID:    sl.ChannelID,
			SubscriberID: sl.SubscriberID + "-different",
		}))
		assert.NotEqual(t, data.ComputeChechsum(sl), data.ComputeChechsum(domain.SubscriberLocator{
			ChannelID:    sl.ChannelID + "-different",
			SubscriberID: sl.SubscriberID,
		}))
	}
}

func TestUnmatchAckHandle(t *testing.T) {
	sl := domain.SubscriberLocator{
		ChannelID:    "ch-1",
		SubscriberID: "sbsc-1",
	}
	data := ackHandleData{
		LastMessageClock: -1234,
		Checksum:         "???",
	}
	data.Checksum = data.ComputeChechsum(sl)
	dataJSON, _ := json.Marshal(data)

	_, err := decodeAckHandle(domain.AckHandle{SubscriberLocator: sl, Handle: string(dataJSON)})
	assert.NoError(t, err) // Must match

	for _, unmatchLocator := range []domain.SubscriberLocator{
		{ChannelID: sl.ChannelID, SubscriberID: "different-subscriber"},
		{ChannelID: "different-channel", SubscriberID: sl.SubscriberID},
	} {
		_, err := decodeAckHandle(domain.AckHandle{SubscriberLocator: unmatchLocator, Handle: string(dataJSON)})
		dspstesting.IsError(t, domain.ErrMalformedAckHandle, err)
		assert.Contains(t, err.Error(), "checksum unmatch")
	}
}

func TestInvalidAckHandle(t *testing.T) {
	_, err := decodeAckHandle(domain.AckHandle{
		SubscriberLocator: domain.SubscriberLocator{
			ChannelID:    "ch-1",
			SubscriberID: "sbsc-1",
		},
		Handle: `{ invalid-json }`,
	})
	dspstesting.IsError(t, domain.ErrMalformedAckHandle, err)
	assert.Contains(t, err.Error(), "JSON parse error")

	_, err = decodeAckHandle(domain.AckHandle{
		SubscriberLocator: domain.SubscriberLocator{
			ChannelID:    "ch-1",
			SubscriberID: "sbsc-1",
		},
		Handle: `{ "clk": "invalid-clock" }`,
	})
	dspstesting.IsError(t, domain.ErrMalformedAckHandle, err)
	assert.Contains(t, err.Error(), "JSON parse error")
	assert.Contains(t, err.Error(), "cannot unmarshal string into Go struct field ackHandleData.clk of type redis.channelClock")

	_, err = decodeAckHandle(domain.AckHandle{
		SubscriberLocator: domain.SubscriberLocator{
			ChannelID:    "ch-1",
			SubscriberID: "sbsc-1",
		},
		Handle: `{ }`,
	})
	dspstesting.IsError(t, domain.ErrMalformedAckHandle, err)
	assert.Contains(t, err.Error(), "checksum unmatch")

	_, err = decodeAckHandle(domain.AckHandle{
		SubscriberLocator: domain.SubscriberLocator{
			ChannelID:    "ch-1",
			SubscriberID: "sbsc-1",
		},
		Handle: `{ "clk": 1234, "xs": "invalid-checksum" }`,
	})
	dspstesting.IsError(t, domain.ErrMalformedAckHandle, err)
	assert.Contains(t, err.Error(), "checksum unmatch")
}
