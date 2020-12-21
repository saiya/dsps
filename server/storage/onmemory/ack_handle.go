package onmemory

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"hash/crc32"

	"github.com/saiya/dsps/server/domain"
	"golang.org/x/xerrors"
)

// EncodeAckHandle encapsle AckHandle
func encodeAckHandle(sl domain.SubscriberLocator, data ackHandleData) domain.AckHandle {
	data.Checksum = data.ComputeChecksum(sl)
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
		return data, xerrors.Errorf("Invalid on-memory AckHandle (%s), JSON parse error: %v (%w)", h.Handle, err, domain.ErrMalformedAckHandle)
	}
	if data.ComputeChecksum(h.SubscriberLocator) != data.Checksum {
		return data, xerrors.Errorf("Corrupted AckHandle (%s), checksum unmatch (%w)", h.Handle, domain.ErrMalformedAckHandle)
	}
	return data, nil
}

// AckHandleData represents decoded (raw) ReceiptHandle
type ackHandleData struct {
	LastMessageID domain.MessageID `json:"mid"`
	Checksum      string           `json:"xs"`
}

func (data ackHandleData) ComputeChecksum(sl domain.SubscriberLocator) string {
	hashBuffer := bytes.Buffer{}
	hashBuffer.WriteString("dsps.storage.on-memory")
	hashBuffer.WriteByte(0x00)
	hashBuffer.WriteString(string(sl.ChannelID))
	hashBuffer.WriteByte(0x00)
	hashBuffer.WriteString(string(sl.SubscriberID))
	hashBuffer.WriteByte(0x00)
	hashBuffer.WriteString(string(data.LastMessageID))

	base64Buffer := bytes.Buffer{}
	binary.Write(&base64Buffer, binary.BigEndian, crc32.ChecksumIEEE(hashBuffer.Bytes())) //nolint:errcheck,gosec
	return base64.RawStdEncoding.EncodeToString(base64Buffer.Bytes())
}
