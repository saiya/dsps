package endpoints_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/domain"
	. "github.com/saiya/dsps/server/domain/mock"
	. "github.com/saiya/dsps/server/http"
	. "github.com/saiya/dsps/server/http/testing"
)

func TestPublishEndpointsWithoutPubSubSupport(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	storage := NewMockStorage(ctrl)
	storage.EXPECT().AsPubSubStorage().Return(nil).AnyTimes()
	storage.EXPECT().AsJwtStorage().Return(nil).AnyTimes()

	chID := "my-channel"
	msgID := "msg-1"
	content := `{"hi":"hello!"}`
	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {
		deps.Storage = storage
	}, func(deps *ServerDependencies, baseURL string) {
		res := DoHTTPRequest(t, "PUT", fmt.Sprintf("%s/channel/%s/message/%s", baseURL, chID, msgID), content)
		AssertErrorResponse(t, res, 501, nil, `No PubSub compatible storage available`)
	})
}

func TestChannelPublishSuccess(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	chID := "my-channel"
	msgID := "test-channel-publish-1"
	content := `{"hi":"hello!"}`
	sl := domain.SubscriberLocator{ChannelID: domain.ChannelID(chID), SubscriberID: "sbsc-1"}
	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {}, func(deps *ServerDependencies, baseURL string) {
		assert.NoError(t, deps.Storage.AsPubSubStorage().NewSubscriber(ctx, sl))

		res := DoHTTPRequest(t, "PUT", fmt.Sprintf("%s/channel/%s/message/%s", baseURL, chID, msgID), content)
		AssertResponseJSON(t, res, 200, map[string]interface{}{
			"channelID": chID,
			"messageID": msgID,
		})

		fetched, _, _, err := deps.Storage.AsPubSubStorage().FetchMessages(ctx, sl, 10, domain.Duration{Duration: 1})
		assert.Equal(t, 1, len(fetched))
		assert.Equal(t, msgID, string(fetched[0].MessageID))
		assert.NoError(t, err)

		// Should be idempotent
		res = DoHTTPRequest(t, "PUT", fmt.Sprintf("%s/channel/%s/message/%s", baseURL, chID, msgID), content)
		AssertResponseJSON(t, res, 200, map[string]interface{}{
			"channelID": chID,
			"messageID": msgID,
		})
	})
}

func TestChannelPublishFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	storage, pubsub, _ := NewMockStorages(ctrl)

	chID := "my-channel"
	msgID := "msg-1"
	content := `{"hi":"hello!"}`
	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {
		deps.Storage = storage
	}, func(deps *ServerDependencies, baseURL string) {
		res := DoHTTPRequest(t, "PUT", fmt.Sprintf("%s/channel/%s/message/%s", baseURL, "** INVALID CHANNEL ID **", msgID), content)
		AssertErrorResponse(t, res, 400, nil, `Invalid "channelID" parameter`)

		res = DoHTTPRequest(t, "PUT", fmt.Sprintf("%s/channel/%s/message/%s", baseURL, chID, "** INVALID MESSAGE ID **"), content)
		AssertErrorResponse(t, res, 400, nil, `Invalid "messageID" parameter`)

		res = DoHTTPRequest(t, "PUT", fmt.Sprintf("%s/channel/%s/message/%s", baseURL, chID, msgID), ``)
		AssertErrorResponse(t, res, 400, nil, `Request body is not JSON`)

		res = DoHTTPRequest(t, "PUT", fmt.Sprintf("%s/channel/%s/message/%s", baseURL, chID, msgID), `{`)
		AssertErrorResponse(t, res, 400, nil, `Request body is not JSON`)

		pubsub.EXPECT().PublishMessages(gomock.Any(), gomock.Any()).Return(domain.ErrInvalidChannel)
		res = DoHTTPRequest(t, "PUT", fmt.Sprintf("%s/channel/%s/message/%s", baseURL, chID, msgID), content)
		AssertErrorResponse(t, res, 403, domain.ErrInvalidChannel, "")

		pubsub.EXPECT().PublishMessages(gomock.Any(), gomock.Any()).Return(errors.New("mock error"))
		res = DoHTTPRequest(t, "PUT", fmt.Sprintf("%s/channel/%s/message/%s", baseURL, chID, msgID), content)
		AssertInternalServerErrorResponse(t, res)
	})
}
