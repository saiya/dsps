package redis

import (
	"encoding/json"

	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/domain"
)

type messageEnvelope struct {
	ID      domain.MessageID `json:"id"`
	Content json.RawMessage  `json:"content"`
}

func wrapMessage(msg domain.Message) (string, error) {
	data, err := json.Marshal(messageEnvelope{
		ID:      msg.MessageID,
		Content: msg.Content,
	})
	if err != nil {
		return "", xerrors.Errorf(`%w: %v`, domain.ErrMalformedMessageJSON, err)
	}
	return string(data), nil
}

func unwrapMessage(ch domain.ChannelID, raw string) (*domain.Message, error) {
	envelope := messageEnvelope{}
	if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
		return nil, xerrors.Errorf(`Failed to parse message envelope JSON '%s': %w`, string(raw), err)
	}
	return &domain.Message{
		MessageLocator: domain.MessageLocator{
			ChannelID: ch,
			MessageID: envelope.ID,
		},
		Content: envelope.Content,
	}, nil
}
