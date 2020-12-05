package redis

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/saiya/dsps/server/domain"
	dspstesting "github.com/saiya/dsps/server/testing"
)

func TestIsOldMessagesMGetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ch := randomChannelID(t)
	keys := keyOfChannel(ch)
	sl := domain.SubscriberLocator{ChannelID: ch, SubscriberID: "sbsc-1"}
	msgs := []domain.MessageLocator{
		{ChannelID: ch, MessageID: "msg-1"},
		{ChannelID: ch, MessageID: "msg-2"},
	}

	s, redisCmd := newMockedRedisStorage(ctrl)
	errToReturn := errors.New("Mocked redis error")
	redisCmd.EXPECT().MGet(
		gomock.Any(),
		keys.Clock(),
		keys.SubscriberCursor(sl.SubscriberID),
		keys.MessageDedup("msg-1"),
		keys.MessageDedup("msg-2"),
	).Return(nil, errToReturn)

	_, err := s.IsOldMessages(context.Background(), sl, msgs)
	dspstesting.IsError(t, errToReturn, err)
}

func TestIsOldMessagesClocksNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ch := randomChannelID(t)
	keys := keyOfChannel(ch)
	sl := domain.SubscriberLocator{ChannelID: ch, SubscriberID: "sbsc-1"}
	msgs := []domain.MessageLocator{
		{ChannelID: ch, MessageID: "msg-1"},
		{ChannelID: ch, MessageID: "msg-2"},
	}

	s, redisCmd := newMockedRedisStorage(ctrl)
	redisCmd.EXPECT().MGet(
		gomock.Any(),
		keys.Clock(),
		keys.SubscriberCursor(sl.SubscriberID),
		keys.MessageDedup("msg-1"),
		keys.MessageDedup("msg-2"),
	).Return(strPList(
		t,
		"INVALID", // Clock
		"INVALID", // SubscriberCursor
		"10",
		"11",
	), nil)

	_, err := s.IsOldMessages(context.Background(), sl, msgs)
	dspstesting.IsError(t, domain.ErrSubscriptionNotFound, err)
}
