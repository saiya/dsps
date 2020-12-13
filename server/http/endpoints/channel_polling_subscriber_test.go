package endpoints_test

import (
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

func TestPollingSubscriberPutSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	chID := "my-channel"
	subscriberID := "sbsc-1"
	content := `{"hi":"hello!"}`
	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {}, func(deps *ServerDependencies, baseURL string) {
		res := DoHTTPRequest(t, "PUT", fmt.Sprintf("%s/channel/%s/subscription/polling/%s", baseURL, chID, subscriberID), content)
		assert.Equal(t, 200, res.StatusCode)
		AssertResponseJSON(t, res, 200, map[string]interface{}{
			"channelID":    chID,
			"subscriberID": subscriberID,
		})
	})
}

func TestPollingSubscriberPutFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	storage := NewMockStorage(ctrl)
	pubsub := NewMockPubSubStorage(ctrl)
	storage.EXPECT().AsPubSubStorage().Return(pubsub).AnyTimes()

	chID := "my-channel"
	subscriberID := "sbsc-1"
	content := `{"hi":"hello!"}`
	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {
		deps.Storage = storage
	}, func(deps *ServerDependencies, baseURL string) {
		res := DoHTTPRequest(t, "PUT", fmt.Sprintf("%s/channel/%s/subscription/polling/%s", baseURL, "*** INVALID ***", subscriberID), content)
		AssertErrorResponse(t, res, 400, nil, `Invalid "channelID" parameter`)

		res = DoHTTPRequest(t, "PUT", fmt.Sprintf("%s/channel/%s/subscription/polling/%s", baseURL, chID, "*** INVALID ***"), content)
		AssertErrorResponse(t, res, 400, nil, `Invalid "subscriberID" parameter`)

		pubsub.EXPECT().NewSubscriber(gomock.Any(), gomock.Any()).Return(domain.ErrInvalidChannel)
		res = DoHTTPRequest(t, "PUT", fmt.Sprintf("%s/channel/%s/subscription/polling/%s", baseURL, chID, subscriberID), content)
		AssertErrorResponse(t, res, 403, domain.ErrInvalidChannel, "")

		pubsub.EXPECT().NewSubscriber(gomock.Any(), gomock.Any()).Return(errors.New("mock error"))
		res = DoHTTPRequest(t, "PUT", fmt.Sprintf("%s/channel/%s/subscription/polling/%s", baseURL, chID, subscriberID), content)
		AssertErrorResponse(t, res, 500, nil, "")
	})
}
