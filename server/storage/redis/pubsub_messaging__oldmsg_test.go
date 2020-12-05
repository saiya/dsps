package redis

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/saiya/dsps/server/domain"
	dspstesting "github.com/saiya/dsps/server/testing"
	"github.com/stretchr/testify/assert"
)

func TestPublishMessagesRedisScriptSingleError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ch := randomChannelID(t)
	msg1 := domain.Message{
		MessageLocator: domain.MessageLocator{
			ChannelID: ch,
			MessageID: "msg-1",
		},
		Content: json.RawMessage(`{}`),
	}
	errToReturn := errors.New("Mocked redis error")

	s, redisCmd := newMockedRedisStorage(ctrl)
	redisCmd.EXPECT().RunScript(gomock.Any(), publishMessageScript, gomock.Any(), gomock.Any()).Return("", errToReturn)
	// No need to call Redis PUBLISH because no message sent.
	redisCmd.EXPECT().Publish(gomock.Any(), s.redisPubSubKeyOf(ch), "new message").MaxTimes(0)

	err := s.PublishMessages(context.Background(), []domain.Message{msg1})
	dspstesting.IsError(t, errToReturn, err)
}

func TestPublishMessagesRedisScriptSuccessAndError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ch := randomChannelID(t)
	msg1 := domain.Message{
		MessageLocator: domain.MessageLocator{
			ChannelID: ch,
			MessageID: "msg-1",
		},
		Content: json.RawMessage(`{}`),
	}
	msg2 := domain.Message{
		MessageLocator: domain.MessageLocator{
			ChannelID: ch,
			MessageID: "msg-2",
		},
		Content: json.RawMessage(`{}`),
	}
	errToReturn := errors.New("Mocked redis error")

	s, redisCmd := newMockedRedisStorage(ctrl)
	firstCall := redisCmd.EXPECT().RunScript(gomock.Any(), publishMessageScript, gomock.Any(), gomock.Any()).Return("OK", nil)
	redisCmd.EXPECT().RunScript(gomock.Any(), publishMessageScript, gomock.Any(), gomock.Any()).Return("", errToReturn).After(firstCall)
	// Redis PUBLISH must be called when one (or more) messages sent.
	redisCmd.EXPECT().Publish(gomock.Any(), s.redisPubSubKeyOf(ch), "new message").Return(nil)

	err := s.PublishMessages(context.Background(), []domain.Message{msg1, msg2})
	dspstesting.IsError(t, errToReturn, err)
}

func TestPublishMessagesRedisPublishError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ch := randomChannelID(t)
	msg1 := domain.Message{
		MessageLocator: domain.MessageLocator{
			ChannelID: ch,
			MessageID: "msg-1",
		},
		Content: json.RawMessage(`{}`),
	}
	errToReturn := errors.New("Mocked redis error")

	s, redisCmd := newMockedRedisStorage(ctrl)
	redisCmd.EXPECT().RunScript(gomock.Any(), publishMessageScript, gomock.Any(), gomock.Any()).Return("OK", nil)
	redisCmd.EXPECT().Publish(gomock.Any(), s.redisPubSubKeyOf(ch), "new message").Return(errToReturn)

	err := s.PublishMessages(context.Background(), []domain.Message{msg1})
	assert.NoError(t, err)
}
