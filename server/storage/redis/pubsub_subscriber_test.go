package redis

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/saiya/dsps/server/domain"
	storagetesting "github.com/saiya/dsps/server/storage/testing"
	dspstesting "github.com/saiya/dsps/server/testing"
	"github.com/stretchr/testify/assert"
)

func TestRemoveSubscriberError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ch := randomChannelID(t)
	sl := domain.SubscriberLocator{ChannelID: ch, SubscriberID: "sbsc-1"}
	keys := keyOfChannel(ch)

	s, redisCmd := newMockedRedisStorage(ctrl)
	errToReturn := errors.New("Mocked redis error")
	redisCmd.EXPECT().Del(gomock.Any(), keys.SubscriberCursor(sl.SubscriberID)).Return(errToReturn)

	err := s.RemoveSubscriber(context.Background(), sl)
	dspstesting.IsError(t, errToReturn, err)
}

func TestExtendSubscriberTTLSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ch := randomChannelID(t)
	sl := domain.SubscriberLocator{ChannelID: ch, SubscriberID: "sbsc-1"}
	keys := keyOfChannel(ch)

	s, redisCmd := newMockedRedisStorage(ctrl)
	redisCmd.EXPECT().Expire(gomock.Any(), keys.Clock(), storagetesting.StubChannelExpire.Duration+ttlMargin).Return(nil)
	redisCmd.EXPECT().Expire(gomock.Any(), keys.SubscriberCursor(sl.SubscriberID), storagetesting.StubChannelExpire.Duration+ttlMargin).Return(nil)

	assert.NoError(t, s.extendSubscriberTTL(context.Background(), sl))
}

func TestExtendSubscriberTTLError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ch := randomChannelID(t)
	sl := domain.SubscriberLocator{ChannelID: ch, SubscriberID: "sbsc-1"}
	keys := keyOfChannel(ch)

	s, redisCmd := newMockedRedisStorage(ctrl)
	errToReturn := errors.New("Mocked redis error")
	redisCmd.EXPECT().Expire(gomock.Any(), keys.Clock(), storagetesting.StubChannelExpire.Duration+ttlMargin).Return(errToReturn)
	redisCmd.EXPECT().Expire(gomock.Any(), keys.SubscriberCursor(sl.SubscriberID), storagetesting.StubChannelExpire.Duration+ttlMargin).Return(errToReturn)

	dspstesting.IsError(t, errToReturn, s.extendSubscriberTTL(context.Background(), sl))
}
