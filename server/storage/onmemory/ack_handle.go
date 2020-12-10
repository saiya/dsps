package onmemory

import (
	"github.com/saiya/dsps/server/domain"
)

// AckHandleData represents decoded (raw) ReceiptHandle
type ackHandleData struct {
	LastMessageID domain.MessageID
}

// EncodeAckHandle encapsle AckHandle
func encodeAckHandle(sl domain.SubscriberLocator, data ackHandleData) domain.AckHandle {
	return domain.AckHandle{
		SubscriberLocator: sl,
		// FIXME: Encode ChannelID and SubscriberID
		Handle: string(data.LastMessageID),
	}
}

// DecodeAckHandle decodes AckHandle
func decodeAckHandle(h domain.AckHandle) (ackHandleData, error) {
	return ackHandleData{
		// FIXME: Decode + Validate with salt (calc by channelID + subscriberID)
		LastMessageID: domain.MessageID(h.Handle),
	}, nil
}
