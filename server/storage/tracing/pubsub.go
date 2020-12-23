package tracing

import (
	"context"

	"github.com/saiya/dsps/server/domain"
)

func (ts *tracingStorage) NewSubscriber(ctx context.Context, sl domain.SubscriberLocator) error {
	ctx, end := ts.t.StartStorageSpan(ctx, ts.id, "NewSubscriber")
	ts.t.SetSubscriberAttributes(ctx, sl)
	defer end()
	return ts.pubsub.NewSubscriber(ctx, sl)
}

func (ts *tracingStorage) RemoveSubscriber(ctx context.Context, sl domain.SubscriberLocator) error {
	ctx, end := ts.t.StartStorageSpan(ctx, ts.id, "RemoveSubscriber")
	ts.t.SetSubscriberAttributes(ctx, sl)
	defer end()
	return ts.pubsub.RemoveSubscriber(ctx, sl)
}

func (ts *tracingStorage) PublishMessages(ctx context.Context, msgs []domain.Message) error {
	ctx, end := ts.t.StartStorageSpan(ctx, ts.id, "PublishMessages")
	defer end()
	return ts.pubsub.PublishMessages(ctx, msgs)
}

func (ts *tracingStorage) FetchMessages(ctx context.Context, sl domain.SubscriberLocator, max int, waituntil domain.Duration) (messages []domain.Message, moreMessages bool, ackHandle domain.AckHandle, err error) {
	ctx, end := ts.t.StartStorageSpan(ctx, ts.id, "FetchMessages")
	ts.t.SetSubscriberAttributes(ctx, sl)
	defer end()
	return ts.pubsub.FetchMessages(ctx, sl, max, waituntil)
}

func (ts *tracingStorage) AcknowledgeMessages(ctx context.Context, handle domain.AckHandle) error {
	ctx, end := ts.t.StartStorageSpan(ctx, ts.id, "AcknowledgeMessages")
	ts.t.SetSubscriberAttributes(ctx, handle.SubscriberLocator)
	defer end()
	return ts.pubsub.AcknowledgeMessages(ctx, handle)
}

func (ts *tracingStorage) IsOldMessages(ctx context.Context, sl domain.SubscriberLocator, msgs []domain.MessageLocator) (map[domain.MessageLocator]bool, error) {
	ctx, end := ts.t.StartStorageSpan(ctx, ts.id, "IsOldMessages")
	ts.t.SetSubscriberAttributes(ctx, sl)
	defer end()
	return ts.pubsub.IsOldMessages(ctx, sl, msgs)
}
