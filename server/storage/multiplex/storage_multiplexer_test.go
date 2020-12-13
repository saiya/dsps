package multiplex_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	. "github.com/saiya/dsps/server/domain/mock"
	. "github.com/saiya/dsps/server/storage/multiplex"
	"github.com/saiya/dsps/server/storage/onmemory"
	. "github.com/saiya/dsps/server/storage/testing"
	. "github.com/saiya/dsps/server/testing"
)

func TestLongPollingEarlyReturn(t *testing.T) {
	ctx := context.Background()
	clock := domain.RealSystemClock
	cp := StubChannelProvider

	s1, err := onmemory.NewOnmemoryStorage(ctx, &config.OnmemoryStorageConfig{}, clock, cp)
	assert.NoError(t, err)
	s2, err := onmemory.NewOnmemoryStorage(ctx, &config.OnmemoryStorageConfig{}, clock, cp)
	assert.NoError(t, err)
	s, err := NewStorageMultiplexer(map[domain.StorageID]domain.Storage{"s1": s1, "s2": s2})
	assert.NoError(t, err)

	// Start subscription
	ch := domain.ChannelID("ch-1")
	sl := domain.SubscriberLocator{ChannelID: ch, SubscriberID: "sbsc-1"}
	assert.NoError(t, s.AsPubSubStorage().NewSubscriber(ctx, sl))

	// Publish only to s1.
	msgs := []domain.Message{
		{
			MessageLocator: domain.MessageLocator{ChannelID: ch, MessageID: "msg-1"},
			Content:        json.RawMessage(`{}`),
		},
		{
			MessageLocator: domain.MessageLocator{ChannelID: ch, MessageID: "msg-2"},
			Content:        json.RawMessage(`{}`),
		},
	}
	assert.NoError(t, s1.AsPubSubStorage().PublishMessages(ctx, msgs))

	// Fetch from both.
	// Multiplexer should return immediately because s1 returns messages instantly.
	ctxD, cancel := context.WithDeadline(ctx, time.Now().Add((300+200)*time.Millisecond))
	defer cancel()
	fetched, _, _, err := s.AsPubSubStorage().FetchMessages(ctxD, sl, 10, MakeDuration("30s"))
	assert.NoError(t, err)
	assert.NoError(t, ctxD.Err()) // FetchMessages must return immediately in this case.
	MessagesEqual(t, msgs, fetched)
}

func TestAckHandleDurability(t *testing.T) {
	ctx := context.Background()
	clock := domain.RealSystemClock
	cp := StubChannelProvider

	s1, err := onmemory.NewOnmemoryStorage(ctx, &config.OnmemoryStorageConfig{}, clock, cp)
	assert.NoError(t, err)
	s2, err := onmemory.NewOnmemoryStorage(ctx, &config.OnmemoryStorageConfig{}, clock, cp)
	assert.NoError(t, err)
	s3, err := onmemory.NewOnmemoryStorage(ctx, &config.OnmemoryStorageConfig{}, clock, cp)
	assert.NoError(t, err)

	// Generate AckHandle of s1 + s2
	sBefore, err := NewStorageMultiplexer(map[domain.StorageID]domain.Storage{"s1": s1, "s2": s2})
	assert.NoError(t, err)
	ch := domain.ChannelID("ch-1")
	sl := domain.SubscriberLocator{ChannelID: ch, SubscriberID: "sbsc-1"}
	assert.NoError(t, sBefore.AsPubSubStorage().NewSubscriber(ctx, sl))
	msgs := []domain.Message{
		{
			MessageLocator: domain.MessageLocator{ChannelID: ch, MessageID: "msg-1"},
			Content:        json.RawMessage(`{}`),
		},
		{
			MessageLocator: domain.MessageLocator{ChannelID: ch, MessageID: "msg-2"},
			Content:        json.RawMessage(`{}`),
		},
	}
	assert.NoError(t, sBefore.AsPubSubStorage().PublishMessages(ctx, msgs))
	fetched, _, ackHandle, err := sBefore.AsPubSubStorage().FetchMessages(ctx, sl, 10, MakeDuration("30s"))
	assert.NoError(t, err)
	MessagesEqual(t, msgs, fetched)

	// Consume AckHandle with s2 + s3
	// Multiplexer must successfully consume handle
	sAfter, err := NewStorageMultiplexer(map[domain.StorageID]domain.Storage{"s2": s2, "s3": s3})
	assert.NoError(t, err)
	assert.NoError(t, sAfter.AsPubSubStorage().AcknowledgeMessages(ctx, ackHandle))
	fetched, _, _, err = sAfter.AsPubSubStorage().FetchMessages(ctx, sl, 10, MakeDuration("10ms"))
	assert.NoError(t, err)
	MessagesEqual(t, []domain.Message{}, fetched)

	// Fetch message from s1 + s2 again.
	// After acknowledgement, old messages should not be return even if one (or more) storages left behind.
	fetched, _, _, err = sBefore.AsPubSubStorage().FetchMessages(ctx, sl, 10, MakeDuration("10ms"))
	assert.NoError(t, err)
	MessagesEqual(t, []domain.Message{}, fetched)
}

func TestSubscriptionDurability(t *testing.T) {
	ctx := context.Background()
	clock := domain.RealSystemClock
	cp := StubChannelProvider

	s1, err := onmemory.NewOnmemoryStorage(ctx, &config.OnmemoryStorageConfig{}, clock, cp)
	assert.NoError(t, err)
	s2, err := onmemory.NewOnmemoryStorage(ctx, &config.OnmemoryStorageConfig{}, clock, cp)
	assert.NoError(t, err)
	s, err := NewStorageMultiplexer(map[domain.StorageID]domain.Storage{"s1": s1, "s2": s2})
	assert.NoError(t, err)

	// Start subscription only on s1
	ch := domain.ChannelID("ch-1")
	sl := domain.SubscriberLocator{ChannelID: ch, SubscriberID: "sbsc-1"}
	assert.NoError(t, s1.AsPubSubStorage().NewSubscriber(ctx, sl))

	// Publish messages (both s1 and s2)
	msgs := []domain.Message{
		{
			MessageLocator: domain.MessageLocator{ChannelID: ch, MessageID: "msg-1"},
			Content:        json.RawMessage(`{}`),
		},
		{
			MessageLocator: domain.MessageLocator{ChannelID: ch, MessageID: "msg-2"},
			Content:        json.RawMessage(`{}`),
		},
	}
	assert.NoError(t, s.AsPubSubStorage().PublishMessages(ctx, msgs))

	// Fetch & Ack from both s1 and s2.
	// Multiplexer should automatically create subscription on s2.
	fetched, _, ackHandle, err := s.AsPubSubStorage().FetchMessages(ctx, sl, 10, MakeDuration("300ms"))
	assert.NoError(t, err)
	MessagesEqual(t, msgs, fetched)
	assert.NoError(t, s.AsPubSubStorage().AcknowledgeMessages(ctx, ackHandle))

	// Publish messages (both s1 and s2)
	msgs = []domain.Message{
		{
			MessageLocator: domain.MessageLocator{ChannelID: ch, MessageID: "msg-3"},
			Content:        json.RawMessage(`{}`),
		},
		{
			MessageLocator: domain.MessageLocator{ChannelID: ch, MessageID: "msg-4"},
			Content:        json.RawMessage(`{}`),
		},
	}
	assert.NoError(t, s.AsPubSubStorage().PublishMessages(ctx, msgs))

	// Delete subscription on s1.
	// Subscription on s2 (automatically created) should still alive.
	assert.NoError(t, s1.AsPubSubStorage().RemoveSubscriber(ctx, sl))

	// Fetch from both s1 and s2.
	// It should work because subscription on s2 alive.
	fetched, _, _, err = s.AsPubSubStorage().FetchMessages(ctx, sl, 10, MakeDuration("300ms"))
	assert.NoError(t, err)
	MessagesEqual(t, msgs, fetched)

	// Delete subscription from both.
	assert.NoError(t, s.AsPubSubStorage().RemoveSubscriber(ctx, sl))

	// Fetch from both s1 and s2.
	// Should fail because no subscription alive on any storages.
	_, _, _, err = s.AsPubSubStorage().FetchMessages(ctx, sl, 10, MakeDuration("300ms"))
	IsError(t, domain.ErrSubscriptionNotFound, err)
}

func TestOldMessageDurability(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ch := domain.ChannelID("ch-1")
	sl := domain.SubscriberLocator{ChannelID: ch, SubscriberID: "sbsc-1"}
	msgs := []domain.MessageLocator{
		{ChannelID: ch, MessageID: "msg-1"},
		{ChannelID: ch, MessageID: "msg-2"},
	}

	s1 := NewMockStorage(ctrl)
	s2 := NewMockStorage(ctrl)
	s1.EXPECT().AsJwtStorage().AnyTimes().Return(nil)
	s2.EXPECT().AsJwtStorage().AnyTimes().Return(nil)
	pubsub := NewMockPubSubStorage(ctrl)
	s1.EXPECT().AsPubSubStorage().AnyTimes().Return(pubsub)
	s2.EXPECT().AsPubSubStorage().AnyTimes().Return(pubsub)
	s, err := NewStorageMultiplexer(map[domain.StorageID]domain.Storage{"s1": s1, "s2": s2})
	assert.NoError(t, err)

	// Multiplexed IsOldMessage() should not fail even if all storages failed.
	pubsub.EXPECT().IsOldMessages(gomock.Any(), sl, msgs).AnyTimes().Return(nil, errors.New("Mock storage error"))
	result, err := s.AsPubSubStorage().IsOldMessages(context.Background(), sl, msgs)
	assert.NoError(t, err)
	assert.Equal(t, map[domain.MessageLocator]bool{
		msgs[0]: false,
		msgs[1]: false,
	}, result)
}

func TestInsufficientStorages(t *testing.T) {
	ctx := context.Background()
	_, err := NewStorageMultiplexer(map[domain.StorageID]domain.Storage{})
	assert.EqualError(t, err, "List of storages must not be empty")

	pubSubDisabledCfg := config.OnmemoryStorageConfig{
		DisablePubSub: true,
	}
	pubSubDisabled1, err := onmemory.NewOnmemoryStorage(ctx, &pubSubDisabledCfg, domain.RealSystemClock, StubChannelProvider)
	assert.NoError(t, err)
	pubSubDisabled2, err := onmemory.NewOnmemoryStorage(ctx, &pubSubDisabledCfg, domain.RealSystemClock, StubChannelProvider)
	assert.NoError(t, err)
	multiWithoutPubSub, err := NewStorageMultiplexer(map[domain.StorageID]domain.Storage{
		"test1": pubSubDisabled1,
		"test2": pubSubDisabled2,
	})
	assert.NoError(t, err)
	assert.Nil(t, multiWithoutPubSub.AsPubSubStorage())
	assert.NotNil(t, multiWithoutPubSub.AsJwtStorage())

	jwtDisabledCfg := config.OnmemoryStorageConfig{
		DisableJwt: true,
	}
	jwtDisabled1, err := onmemory.NewOnmemoryStorage(ctx, &jwtDisabledCfg, domain.RealSystemClock, StubChannelProvider)
	assert.NoError(t, err)
	jwtDisabled2, err := onmemory.NewOnmemoryStorage(ctx, &jwtDisabledCfg, domain.RealSystemClock, StubChannelProvider)
	assert.NoError(t, err)
	multiWithoutJwt, err := NewStorageMultiplexer(map[domain.StorageID]domain.Storage{
		"test1": jwtDisabled1,
		"test2": jwtDisabled2,
	})
	assert.NoError(t, err)
	assert.NotNil(t, multiWithoutJwt.AsPubSubStorage())
	assert.Nil(t, multiWithoutJwt.AsJwtStorage())
}

func TestGetNoFilePressure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock1 := NewMockStorage(ctrl)
	mock1.EXPECT().GetNoFilePressure().Return(21)
	mock1.EXPECT().AsPubSubStorage().Return(nil).AnyTimes()
	mock1.EXPECT().AsJwtStorage().Return(nil).AnyTimes()
	mock2 := NewMockStorage(ctrl)
	mock2.EXPECT().GetNoFilePressure().Return(300)
	mock2.EXPECT().AsPubSubStorage().Return(nil).AnyTimes()
	mock2.EXPECT().AsJwtStorage().Return(nil).AnyTimes()

	s, err := NewStorageMultiplexer(map[domain.StorageID]domain.Storage{
		"mock1": mock1,
		"mock2": mock2,
	})
	assert.NoError(t, err)
	assert.Equal(t, 321, s.GetNoFilePressure())
}
