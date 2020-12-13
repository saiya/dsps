package endpoints_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/saiya/dsps/server/domain"
	. "github.com/saiya/dsps/server/domain/mock"
	. "github.com/saiya/dsps/server/http"
	. "github.com/saiya/dsps/server/http/testing"
)

func TestChannelPublishSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	chID := "my-channel"
	msgID := "msg-1"
	content := `{"hi":"hello!"}`
	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {}, func(deps *ServerDependencies, baseURL string) {
		res := DoHTTPRequest(t, "PUT", fmt.Sprintf("%s/channel/%s/message/%s", baseURL, chID, msgID), content)
		AssertResponseJSON(t, res, 200, map[string]interface{}{
			"channelID": chID,
			"messageID": msgID,
		})
	})
}

func TestChannelPublishFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	storage := NewMockStorage(ctrl)
	pubsub := NewMockPubSubStorage(ctrl)
	storage.EXPECT().AsPubSubStorage().Return(pubsub).AnyTimes()

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
		AssertErrorResponse(t, res, 500, nil, "")
	})
}
