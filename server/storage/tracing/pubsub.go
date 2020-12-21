package tracing

import (
	"context"

	"github.com/saiya/dsps/server/domain"
)

func (ts *tracingStorage) NewSubscriber(ctx context.Context, sl domain.SubscriberLocator) error {
	return ts.pubsub.NewSubscriber(ctx, sl)
}

func (ts *tracingStorage) RemoveSubscriber(ctx context.Context, sl domain.SubscriberLocator) error {
	return ts.pubsub.RemoveSubscriber(ctx, sl)
}

func (ts *tracingStorage) PublishMessages(ctx context.Context, msgs []domain.Message) error {
	return ts.pubsub.PublishMessages(ctx, msgs)
}

func (ts *tracingStorage) FetchMessages(ctx context.Context, sl domain.SubscriberLocator, max int, waituntil domain.Duration) (messages []domain.Message, moreMessages bool, ackHandle domain.AckHandle, err error) {
	return ts.pubsub.FetchMessages(ctx, sl, max, waituntil)
}

func (ts *tracingStorage) AcknowledgeMessages(ctx context.Context, handle domain.AckHandle) error {
	return ts.pubsub.AcknowledgeMessages(ctx, handle)
}

func (ts *tracingStorage) IsOldMessages(ctx context.Context, sl domain.SubscriberLocator, msgs []domain.MessageLocator) (map[domain.MessageLocator]bool, error) {
	return ts.pubsub.IsOldMessages(ctx, sl, msgs)
}
