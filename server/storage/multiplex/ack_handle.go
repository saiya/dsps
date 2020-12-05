package multiplex

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/saiya/dsps/server/domain"
)

// EncodeMultiplexAckHandle encapsle AckHandle of multiplex storage
func encodeMultiplexAckHandle(handles map[domain.StorageID]domain.AckHandle) (domain.AckHandle, error) {
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
func decodeMultiplexAckHandle(h domain.AckHandle) (map[domain.StorageID]domain.AckHandle, error) {
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
