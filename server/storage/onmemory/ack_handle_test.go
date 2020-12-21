package onmemory

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
		{LastMessageID: "msg-1"},
	} {
		data.Checksum = "INVALID"
		data.ComputeChecksum(sl)
		assert.Equal(t, "INVALID", data.Checksum, "ComputeChechsum() should not modify struct")

		assert.Equal(t, data.ComputeChecksum(sl), data.ComputeChecksum(sl))
		assert.NotEqual(t, data.ComputeChecksum(sl), ackHandleData{LastMessageID: data.LastMessageID + "-diff"}.ComputeChecksum(sl))
		assert.NotEqual(t, data.ComputeChecksum(sl), data.ComputeChecksum(domain.SubscriberLocator{
			ChannelID:    sl.ChannelID,
			SubscriberID: sl.SubscriberID + "-different",
		}))
		assert.NotEqual(t, data.ComputeChecksum(sl), data.ComputeChecksum(domain.SubscriberLocator{
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
		LastMessageID: "msg-1",
		Checksum:      "???",
	}
	data.Checksum = data.ComputeChecksum(sl)
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
		Handle: `{ "mid": 1234 }`,
	})
	dspstesting.IsError(t, domain.ErrMalformedAckHandle, err)
	assert.Contains(t, err.Error(), "JSON parse error")
	assert.Contains(t, err.Error(), "cannot unmarshal number into Go struct field ackHandleData.mid of type domain.MessageID ")

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
		Handle: `{ "mid": "invalid-message-id", "xs": "invalid-checksum" }`,
	})
	dspstesting.IsError(t, domain.ErrMalformedAckHandle, err)
	assert.Contains(t, err.Error(), "checksum unmatch")
}
