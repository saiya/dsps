package outgoing

import (
	"context"
	"encoding/json"

	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/domain"
)

// See server/doc/outgoing-webhook.md
type outgoingWebhookBody struct {
	Type      string          `json:"type"`
	ChannelID string          `json:"channelID"`
	MessageID string          `json:"messageID"`
	Content   json.RawMessage `json:"content"`
}

func encodeWebhookBody(ctx context.Context, msg domain.Message) (string, error) {
	body := outgoingWebhookBody{
		Type:      "dsps.channel.outgoing-webhook",
		ChannelID: string(msg.ChannelID),
		MessageID: string(msg.MessageID),
		Content:   msg.Content,
	}
	bytes, err := json.Marshal(body)
	if err != nil {
		return "", xerrors.Errorf(`failed to make request body of outgoing-webhook (channelID: %s, messageID: %s): %w`, body.ChannelID, body.MessageID, err)
	}
	return string(bytes), nil
}
