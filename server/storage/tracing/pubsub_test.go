package tracing_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/saiya/dsps/server/domain"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

func TestPubSubTrace(t *testing.T) {
	chID := "ch-1"
	sbscID := "sbsc-1"
	msgID := "msg-1"
	msgLocator := domain.MessageLocator{ChannelID: domain.ChannelID(chID), MessageID: domain.MessageID(msgID)}
	sl := domain.SubscriberLocator{ChannelID: domain.ChannelID(chID), SubscriberID: domain.SubscriberID(sbscID)}
	tr := testTracing(t, func(s domain.Storage) {
		ctx := context.Background()
		pubsub := s.AsPubSubStorage()

		assert.NoError(t, pubsub.NewSubscriber(ctx, sl))
		assert.NoError(t, pubsub.PublishMessages(ctx, []domain.Message{{MessageLocator: msgLocator, Content: json.RawMessage("{}")}}))
		_, _, ackHandle, err := pubsub.FetchMessages(ctx, sl, 1, domain.Duration{Duration: 100 * time.Millisecond})
		assert.NoError(t, err)
		assert.NoError(t, pubsub.AcknowledgeMessages(ctx, ackHandle))
		_, err = pubsub.IsOldMessages(ctx, sl, []domain.MessageLocator{msgLocator})
		assert.NoError(t, err)
		assert.NoError(t, pubsub.RemoveSubscriber(ctx, sl))
	})
	tr.OT.AssertSpan(0, trace.SpanKindInternal, "DSPS storage NewSubscriber", map[string]interface{}{
		"dsps.storage.id":       "test",
		"messaging.system":      "dsps",
		"messaging.destination": chID,
		"dsps.subscriber_id":    sbscID,
	})
	tr.OT.AssertSpan(1, trace.SpanKindInternal, "DSPS storage PublishMessages", map[string]interface{}{
		"dsps.storage.id": "test",
	})
	tr.OT.AssertSpan(2, trace.SpanKindInternal, "DSPS storage FetchMessages", map[string]interface{}{
		"dsps.storage.id":       "test",
		"messaging.system":      "dsps",
		"messaging.destination": chID,
		"dsps.subscriber_id":    sbscID,
	})
	tr.OT.AssertSpan(3, trace.SpanKindInternal, "DSPS storage AcknowledgeMessages", map[string]interface{}{
		"dsps.storage.id":       "test",
		"messaging.system":      "dsps",
		"messaging.destination": chID,
		"dsps.subscriber_id":    sbscID,
	})
	tr.OT.AssertSpan(4, trace.SpanKindInternal, "DSPS storage IsOldMessages", map[string]interface{}{
		"dsps.storage.id":       "test",
		"messaging.system":      "dsps",
		"messaging.destination": chID,
		"dsps.subscriber_id":    sbscID,
	})
	tr.OT.AssertSpan(5, trace.SpanKindInternal, "DSPS storage RemoveSubscriber", map[string]interface{}{
		"dsps.storage.id":       "test",
		"messaging.system":      "dsps",
		"messaging.destination": chID,
		"dsps.subscriber_id":    sbscID,
	})
	tr.OT.AssertSpan(6, trace.SpanKindInternal, "DSPS storage Shutdown", map[string]interface{}{
		"dsps.storage.id": "test",
	})
}
