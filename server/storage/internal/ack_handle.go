package internal

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/saiya/dsps/server/domain"
)

// AckHandleData represents decoded (raw) ReceiptHandle
type AckHandleData struct {
	LastMessageID domain.MessageID
}

// EncodeAckHandle encapsle AckHandle
func EncodeAckHandle(sl domain.SubscriberLocator, data AckHandleData) domain.AckHandle {
	return domain.AckHandle{
		SubscriberLocator: sl,
		// FIXME: Encode ChannelID and SubscriberID
		Handle: string(data.LastMessageID),
	}
}

// DecodeAckHandle decodes AckHandle
func DecodeAckHandle(h domain.AckHandle) (AckHandleData, error) {
	return AckHandleData{
		// FIXME: Decode + Validate with salt (calc by channelID + subscriberID)
		LastMessageID: domain.MessageID(h.Handle),
	}, nil
}

// EncodeMultiplexAckHandle encapsle AckHandle of multiplex storage
func EncodeMultiplexAckHandle(handles map[domain.StorageID]domain.AckHandle) (domain.AckHandle, error) {
	var sl domain.SubscriberLocator
	raw := map[domain.StorageID]string{}
	for id, h := range handles {
		if (sl == domain.SubscriberLocator{}) {
			sl = h.SubscriberLocator
		} else if sl != h.SubscriberLocator {
			return domain.AckHandle{}, fmt.Errorf("[BUG] Inconsistent SubscriberLocator %v, %v", sl, h.SubscriberLocator)
		}
		raw[id] = h.Handle
	}

	jsonStr, err := json.Marshal(raw)
	if err != nil {
		return domain.AckHandle{}, fmt.Errorf("Failed to encode MultiplexAckHandle (%v): %w", raw, err)
	}
	return domain.AckHandle{
		SubscriberLocator: sl,
		Handle:            base64.StdEncoding.EncodeToString(jsonStr),
	}, nil
}

// DecodeMultiplexAckHandle decodes AckHandle of multiplex storage
func DecodeMultiplexAckHandle(h domain.AckHandle) (map[domain.StorageID]domain.AckHandle, error) {
	jsonStr, err := base64.StdEncoding.DecodeString(h.Handle)
	if err != nil {
		return nil, fmt.Errorf("Failed to base64 decode MultiplexAckHandle \"%s\": %v (%w)", h.Handle, err, domain.ErrMalformedAckHandle)
	}
	raw := map[domain.StorageID]string{}
	if err := json.Unmarshal(jsonStr, &raw); err != nil {
		return nil, fmt.Errorf("Failed to JSON decode MultiplexAckHandle \"%s\": %v (%w)", h.Handle, err, domain.ErrMalformedAckHandle)
	}

	result := map[domain.StorageID]domain.AckHandle{}
	for id, subH := range raw {
		result[id] = domain.AckHandle{
			SubscriberLocator: h.SubscriberLocator,
			Handle:            subH,
		}
	}
	return result, nil
}
