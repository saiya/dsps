package redis

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/saiya/dsps/server/domain"
	dspstesting "github.com/saiya/dsps/server/testing"
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
