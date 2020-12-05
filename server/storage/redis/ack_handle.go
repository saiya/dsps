package redis

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"hash/crc32"

	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/domain"
)

// AckHandleData represents decoded (raw) ReceiptHandle
type ackHandleData struct {
	LastMessageClock channelClock `json:"clk"`
	Checksum         string       `json:"xs"`
}

// Note this method does NOT read nor write ackHandleData.Checksum field.
func (data ackHandleData) ComputeChechsum(sl domain.SubscriberLocator) string {
	hashBuffer := bytes.Buffer{}
	hashBuffer.WriteString("dsps.storage.redis")
	hashBuffer.WriteByte(0x00)
	hashBuffer.WriteString(string(sl.ChannelID))
	hashBuffer.WriteByte(0x00)
	hashBuffer.WriteString(string(sl.SubscriberID))
	hashBuffer.WriteByte(0x00)
	binary.Write(&hashBuffer, binary.BigEndian, data.LastMessageClock) //nolint:errcheck,gosec

	base64Buffer := bytes.Buffer{}
	binary.Write(&base64Buffer, binary.BigEndian, crc32.ChecksumIEEE(hashBuffer.Bytes())) //nolint:errcheck,gosec
	return base64.RawStdEncoding.EncodeToString(base64Buffer.Bytes())
}

// EncodeAckHandle encapsulate AckHandle
func encodeAckHandle(sl domain.SubscriberLocator, data ackHandleData) domain.AckHandle {
	data.Checksum = data.ComputeChechsum(sl)
	encoded, err := json.Marshal(data)
	if err != nil { // Must success
		panic(xerrors.Errorf("Failed to encode Redis ackHandleData (%v): %w", data, err))
	}
	return domain.AckHandle{SubscriberLocator: sl, Handle: string(encoded)}
}

// DecodeAckHandle decodes AckHandle
func decodeAckHandle(h domain.AckHandle) (ackHandleData, error) {
	data := ackHandleData{}
	if err := json.Unmarshal([]byte(h.Handle), &data); err != nil {
		return data, xerrors.Errorf("Invalid Redis AckHandle (%s), JSON parse error: %v (%w)", h.Handle, err, domain.ErrMalformedAckHandle)
	}
	if data.ComputeChechsum(h.SubscriberLocator) != data.Checksum {
		return data, xerrors.Errorf("Corrupted AckHandle (%s), checksum unmatch (%w)", h.Handle, domain.ErrMalformedAckHandle)
	}
	return data, nil
}
