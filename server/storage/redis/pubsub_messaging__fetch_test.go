package redis

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/golang/mock/gomock"
	"github.com/saiya/dsps/server/domain"
	storagetesting "github.com/saiya/dsps/server/storage/testing"
	dspstesting "github.com/saiya/dsps/server/testing"
)

func TestFetchMessagesFirstPollingClockGetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ch := randomChannelID(t)
	keys := keyOfChannel(ch)
	sl := domain.SubscriberLocator{ChannelID: ch, SubscriberID: "sbsc-1"}

	s, redisCmd, _ := newMockedRedisStorageAndPubSubDispatcher(ctrl)

	// (1st fetchMessagesNow) MGET clock cursor
	errToReturn := errors.New(`Mocked Redis error`)
	redisCmd.EXPECT().MGet(gomock.Any(), keys.Clock(), keys.SubscriberCursor(sl.SubscriberID)).Return(nil, errToReturn)

	_, _, _, err := s.FetchMessages(context.Background(), sl, 100, domain.Duration{Duration: 30 * time.Second})
	dspstesting.IsError(t, errToReturn, err)
	assert.Contains(t, err.Error(), "FetchMessages failed due to Redis error (cursor MGET error)")
}

func TestFetchMessagesFirstPollingClockGetInvalidValues(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ch := randomChannelID(t)
	keys := keyOfChannel(ch)
	sl := domain.SubscriberLocator{ChannelID: ch, SubscriberID: "sbsc-1"}

	s, redisCmd, _ := newMockedRedisStorageAndPubSubDispatcher(ctrl)

	// (1st fetchMessagesNow) MGET clock cursor
	redisCmd.EXPECT().MGet(gomock.Any(), keys.Clock(), keys.SubscriberCursor(sl.SubscriberID)).Return(strPList(t, "INVALID", "INVALID"), nil)

	_, _, _, err := s.FetchMessages(context.Background(), sl, 100, domain.Duration{Duration: 30 * time.Second})
	dspstesting.IsError(t, domain.ErrSubscriptionNotFound, err)
}

func TestFetchMessagesFirstPollingGetMsgBodyError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ch := randomChannelID(t)
	keys := keyOfChannel(ch)
	sl := domain.SubscriberLocator{ChannelID: ch, SubscriberID: "sbsc-1"}

	s, redisCmd, _ := newMockedRedisStorageAndPubSubDispatcher(ctrl)

	// (1st fetchMessagesNow) MGET clock cursor
	clocksMget := redisCmd.EXPECT().MGet(gomock.Any(), keys.Clock(), keys.SubscriberCursor(sl.SubscriberID)).Return(strPList(t, "12", "10"), nil)
	redisCmd.EXPECT().Expire(gomock.Any(), keys.Clock(), storagetesting.StubChannelExpire.Duration+ttlMargin).Return(nil)
	redisCmd.EXPECT().Expire(gomock.Any(), keys.SubscriberCursor(sl.SubscriberID), storagetesting.StubChannelExpire.Duration+ttlMargin).Return(nil)
	// (1st fetchMessagesNow) MGET msg1body msg2body
	errorToReturn := errors.New("Mocked redis error")
	redisCmd.EXPECT().MGet(gomock.Any(), keys.MessageBody(11), keys.MessageBody(12)).Return(nil, errorToReturn).After(clocksMget)

	_, _, _, err := s.FetchMessages(context.Background(), sl, 100, domain.Duration{Duration: 3 * time.Second})
	dspstesting.IsError(t, errorToReturn, err)
	assert.Contains(t, err.Error(), "FetchMessages failed due to Redis error (msg MGET error)")
}

func TestFetchMessagesFirstPollingCorruptedMsgBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ch := randomChannelID(t)
	keys := keyOfChannel(ch)
	sl := domain.SubscriberLocator{ChannelID: ch, SubscriberID: "sbsc-1"}

	s, redisCmd, _ := newMockedRedisStorageAndPubSubDispatcher(ctrl)

	// (1st fetchMessagesNow) MGET clock cursor
	clocksMget := redisCmd.EXPECT().MGet(gomock.Any(), keys.Clock(), keys.SubscriberCursor(sl.SubscriberID)).Return(strPList(t, "13", "10"), nil)
	redisCmd.EXPECT().Expire(gomock.Any(), keys.Clock(), storagetesting.StubChannelExpire.Duration+ttlMargin).Return(nil)
	redisCmd.EXPECT().Expire(gomock.Any(), keys.SubscriberCursor(sl.SubscriberID), storagetesting.StubChannelExpire.Duration+ttlMargin).Return(nil)
	// (1st fetchMessagesNow) MGET msg1body msg2body
	msgBody1 := json.RawMessage(`{"hi":"hello1"}`)
	envelope1, _ := json.Marshal(messageEnvelope{ID: "msg1", Content: msgBody1})
	msgBody3 := json.RawMessage(`{"hi":"hello2"}`)
	envelope3, _ := json.Marshal(messageEnvelope{ID: "msg2", Content: msgBody3})
	redisCmd.EXPECT().MGet(gomock.Any(), keys.MessageBody(11), keys.MessageBody(12), keys.MessageBody(13)).Return(strPList(t, string(envelope1), `INVALID JSON`, string(envelope3)), nil).After(clocksMget)

	fetchedMsgs, _, _, err := s.FetchMessages(context.Background(), sl, 100, domain.Duration{Duration: 3 * time.Second})
	assert.NoError(t, err)
	assert.Equal(t, msgBody1, fetchedMsgs[0].Content)
	assert.Equal(t, msgBody3, fetchedMsgs[1].Content)
}

func TestFetchMessagesSecondPollingError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ch := randomChannelID(t)
	keys := keyOfChannel(ch)
	sl := domain.SubscriberLocator{ChannelID: ch, SubscriberID: "sbsc-1"}

	s, redisCmd, dispatcher := newMockedRedisStorageAndPubSubDispatcher(ctrl)

	// (1st fetchMessagesNow) MGET clock cursor
	clocksMget1 := redisCmd.EXPECT().MGet(gomock.Any(), keys.Clock(), keys.SubscriberCursor(sl.SubscriberID)).Return(strPList(t, "10", "10"), nil)
	redisCmd.EXPECT().Expire(gomock.Any(), keys.Clock(), storagetesting.StubChannelExpire.Duration+ttlMargin).Return(nil)
	redisCmd.EXPECT().Expire(gomock.Any(), keys.SubscriberCursor(sl.SubscriberID), storagetesting.StubChannelExpire.Duration+ttlMargin).Return(nil)
	// (1st fetchMessagesNow) MGET (no messages)
	bodyMget1 := redisCmd.EXPECT().MGet(gomock.Any()).Return(nil, nil).Do(func(ctx context.Context, keys ...string) {
		dispatcher.Resolve(s.redisPubSubKeyOf(ch))
	}).After(clocksMget1)

	// (2nd fetchMessagesNow) MGET clock cursor
	errorToReturn := errors.New("Mocked redis error")
	redisCmd.EXPECT().MGet(gomock.Any(), keys.Clock(), keys.SubscriberCursor(sl.SubscriberID)).Return(nil, errorToReturn).After(bodyMget1)

	_, _, _, err := s.FetchMessages(context.Background(), sl, 100, domain.Duration{Duration: 3 * time.Second})
	dspstesting.IsError(t, errorToReturn, err)
	assert.Contains(t, err.Error(), "FetchMessages failed due to Redis error (cursor MGET error)")
}

func TestFetchMessagesSpuriousWakeup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ch := randomChannelID(t)
	keys := keyOfChannel(ch)
	sl := domain.SubscriberLocator{ChannelID: ch, SubscriberID: "sbsc-1"}

	s, redisCmd, dispatcher := newMockedRedisStorageAndPubSubDispatcher(ctrl)

	// (1st fetchMessagesNow) MGET clock cursor
	clocksMget1 := redisCmd.EXPECT().MGet(gomock.Any(), keys.Clock(), keys.SubscriberCursor(sl.SubscriberID)).Return(strPList(t, "10", "10"), nil)
	redisCmd.EXPECT().Expire(gomock.Any(), keys.Clock(), storagetesting.StubChannelExpire.Duration+ttlMargin).Return(nil)
	redisCmd.EXPECT().Expire(gomock.Any(), keys.SubscriberCursor(sl.SubscriberID), storagetesting.StubChannelExpire.Duration+ttlMargin).Return(nil)
	// (1st fetchMessagesNow) MGET (no messages)
	bodyMget1 := redisCmd.EXPECT().MGet(gomock.Any()).Return(nil, nil).Do(func(ctx context.Context, keys ...string) {
		dispatcher.Resolve(s.redisPubSubKeyOf(ch)) // spurious wakeup
	}).After(clocksMget1)

	// (2nd fetchMessagesNow) MGET clock cursor
	clocksMget2 := redisCmd.EXPECT().MGet(gomock.Any(), keys.Clock(), keys.SubscriberCursor(sl.SubscriberID)).Return(strPList(t, "10", "10"), nil).After(bodyMget1)
	redisCmd.EXPECT().Expire(gomock.Any(), keys.Clock(), storagetesting.StubChannelExpire.Duration+ttlMargin).Return(nil).After(bodyMget1)
	redisCmd.EXPECT().Expire(gomock.Any(), keys.SubscriberCursor(sl.SubscriberID), storagetesting.StubChannelExpire.Duration+ttlMargin).Return(nil).After(bodyMget1)
	// (2nd fetchMessagesNow) MGET (no messages)
	redisCmd.EXPECT().MGet(gomock.Any()).Return(nil, nil).After(clocksMget2)

	_, _, _, err := s.FetchMessages(context.Background(), sl, 100, domain.Duration{Duration: 50 * time.Millisecond})
	assert.NoError(t, err)
}

func TestFetchMessagesTTLExtensionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ch := randomChannelID(t)
	keys := keyOfChannel(ch)
	sl := domain.SubscriberLocator{ChannelID: ch, SubscriberID: "sbsc-1"}

	s, redisCmd, dispatcher := newMockedRedisStorageAndPubSubDispatcher(ctrl)

	// (1st fetchMessagesNow) MGET clock cursor
	clocksMget1 := redisCmd.EXPECT().MGet(gomock.Any(), keys.Clock(), keys.SubscriberCursor(sl.SubscriberID)).Return(strPList(t, "10", "10"), nil)
	errToReturn := errors.New("Mocked redis error of EXPIRE command")
	redisCmd.EXPECT().Expire(gomock.Any(), keys.Clock(), storagetesting.StubChannelExpire.Duration+ttlMargin).Return(errToReturn)
	redisCmd.EXPECT().Expire(gomock.Any(), keys.SubscriberCursor(sl.SubscriberID), storagetesting.StubChannelExpire.Duration+ttlMargin).Return(errToReturn)
	// (1st fetchMessagesNow) MGET (no messages)
	redisCmd.EXPECT().MGet(gomock.Any()).Return(nil, nil).Do(func(ctx context.Context, keys ...string) {
		dispatcher.Resolve(s.redisPubSubKeyOf(ch)) // spurious wakeup
	}).After(clocksMget1)

	_, _, _, err := s.FetchMessages(context.Background(), sl, 100, domain.Duration{Duration: 0})
	assert.NoError(t, err)
}
