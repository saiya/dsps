package endpoints_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/domain"
	. "github.com/saiya/dsps/server/domain/mock"
	. "github.com/saiya/dsps/server/http"
	. "github.com/saiya/dsps/server/http/testing"
)

func TestPollingEndpointsWithoutPubSubSupport(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	storage := NewMockStorage(ctrl)
	storage.EXPECT().AsPubSubStorage().Return(nil).AnyTimes()
	storage.EXPECT().AsJwtStorage().Return(nil).AnyTimes()

	sl := domain.SubscriberLocator{
		ChannelID:    "my-channel",
		SubscriberID: "sbsc-1",
	}
	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {
		deps.Storage = storage
	}, func(deps *ServerDependencies, baseURL string) {
		res := DoHTTPRequest(t, "PUT", fmt.Sprintf("%s/channel/%s/subscription/polling/%s", baseURL, sl.ChannelID, sl.SubscriberID), ``)
		AssertErrorResponse(t, res, 501, nil, `No PubSub compatible storage available`)

		res = DoHTTPRequest(t, "DELETE", fmt.Sprintf("%s/channel/%s/subscription/polling/%s", baseURL, sl.ChannelID, sl.SubscriberID), ``)
		AssertErrorResponse(t, res, 501, nil, `No PubSub compatible storage available`)

		res = DoHTTPRequest(t, "GET", fmt.Sprintf("%s/channel/%s/subscription/polling/%s?timeout=0ms", baseURL, sl.ChannelID, sl.SubscriberID), ``)
		AssertErrorResponse(t, res, 501, nil, `No PubSub compatible storage available`)

		res = DoHTTPRequest(t, "DELETE", fmt.Sprintf("%s/channel/%s/subscription/polling/%s/message?ackHandle=%s", baseURL, sl.ChannelID, sl.SubscriberID, "dummy-ack-handle"), ``)
		AssertErrorResponse(t, res, 501, nil, `No PubSub compatible storage available`)
	})
}

func TestPollingSubscriberPutSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sl := domain.SubscriberLocator{
		ChannelID:    "my-channel",
		SubscriberID: "sbsc-1",
	}
	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {}, func(deps *ServerDependencies, baseURL string) {
		res := DoHTTPRequest(t, "PUT", fmt.Sprintf("%s/channel/%s/subscription/polling/%s", baseURL, sl.ChannelID, sl.SubscriberID), ``)
		AssertResponseJSON(t, res, 200, map[string]interface{}{
			"channelID":    string(sl.ChannelID),
			"subscriberID": string(sl.SubscriberID),
		})

		// Should be idempotent
		res = DoHTTPRequest(t, "PUT", fmt.Sprintf("%s/channel/%s/subscription/polling/%s", baseURL, sl.ChannelID, sl.SubscriberID), ``)
		AssertResponseJSON(t, res, 200, map[string]interface{}{
			"channelID":    string(sl.ChannelID),
			"subscriberID": string(sl.SubscriberID),
		})
	})
}

func TestPollingSubscriberPutFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	storage, pubsub, _ := NewMockStorages(ctrl)

	sl := domain.SubscriberLocator{
		ChannelID:    "my-channel",
		SubscriberID: "sbsc-1",
	}
	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {
		deps.Storage = storage
	}, func(deps *ServerDependencies, baseURL string) {
		res := DoHTTPRequest(t, "PUT", fmt.Sprintf("%s/channel/%s/subscription/polling/%s", baseURL, "*** INVALID ***", sl.SubscriberID), ``)
		AssertErrorResponse(t, res, 400, nil, `Invalid "channelID" parameter`)

		res = DoHTTPRequest(t, "PUT", fmt.Sprintf("%s/channel/%s/subscription/polling/%s", baseURL, sl.ChannelID, "*** INVALID ***"), ``)
		AssertErrorResponse(t, res, 400, nil, `Invalid "subscriberID" parameter`)

		pubsub.EXPECT().NewSubscriber(gomock.Any(), sl).Return(domain.ErrInvalidChannel)
		res = DoHTTPRequest(t, "PUT", fmt.Sprintf("%s/channel/%s/subscription/polling/%s", baseURL, sl.ChannelID, sl.SubscriberID), ``)
		AssertErrorResponse(t, res, 403, domain.ErrInvalidChannel, "")

		pubsub.EXPECT().NewSubscriber(gomock.Any(), sl).Return(errors.New("mock error"))
		res = DoHTTPRequest(t, "PUT", fmt.Sprintf("%s/channel/%s/subscription/polling/%s", baseURL, sl.ChannelID, sl.SubscriberID), ``)
		AssertInternalServerErrorResponse(t, res)
	})
}

func TestPollingSubscriberDeleteSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sl := domain.SubscriberLocator{
		ChannelID:    "my-channel",
		SubscriberID: "sbsc-1",
	}
	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {}, func(deps *ServerDependencies, baseURL string) {
		// Success even if not exists
		res := DoHTTPRequest(t, "DELETE", fmt.Sprintf("%s/channel/%s/subscription/polling/%s", baseURL, sl.ChannelID, sl.SubscriberID), ``)
		AssertResponseJSON(t, res, 200, map[string]interface{}{
			"channelID":    string(sl.ChannelID),
			"subscriberID": string(sl.SubscriberID),
		})

		assert.NoError(t, deps.Storage.AsPubSubStorage().NewSubscriber(context.Background(), sl))

		res = DoHTTPRequest(t, "DELETE", fmt.Sprintf("%s/channel/%s/subscription/polling/%s", baseURL, sl.ChannelID, sl.SubscriberID), ``)
		AssertResponseJSON(t, res, 200, map[string]interface{}{
			"channelID":    string(sl.ChannelID),
			"subscriberID": string(sl.SubscriberID),
		})
	})
}

func TestPollingSubscriberDeleteFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	storage, pubsub, _ := NewMockStorages(ctrl)

	sl := domain.SubscriberLocator{
		ChannelID:    "my-channel",
		SubscriberID: "sbsc-1",
	}
	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {
		deps.Storage = storage
	}, func(deps *ServerDependencies, baseURL string) {
		res := DoHTTPRequest(t, "DELETE", fmt.Sprintf("%s/channel/%s/subscription/polling/%s", baseURL, "*** INVALID ***", sl.SubscriberID), ``)
		AssertErrorResponse(t, res, 400, nil, `Invalid "channelID" parameter`)

		res = DoHTTPRequest(t, "DELETE", fmt.Sprintf("%s/channel/%s/subscription/polling/%s", baseURL, sl.ChannelID, "*** INVALID ***"), ``)
		AssertErrorResponse(t, res, 400, nil, `Invalid "subscriberID" parameter`)

		pubsub.EXPECT().RemoveSubscriber(gomock.Any(), sl).Return(domain.ErrInvalidChannel)
		res = DoHTTPRequest(t, "DELETE", fmt.Sprintf("%s/channel/%s/subscription/polling/%s", baseURL, sl.ChannelID, sl.SubscriberID), ``)
		AssertErrorResponse(t, res, 403, domain.ErrInvalidChannel, "")

		pubsub.EXPECT().RemoveSubscriber(gomock.Any(), sl).Return(errors.New("mock error"))
		res = DoHTTPRequest(t, "DELETE", fmt.Sprintf("%s/channel/%s/subscription/polling/%s", baseURL, sl.ChannelID, sl.SubscriberID), ``)
		AssertInternalServerErrorResponse(t, res)
	})
}

func TestPollingSubscriberGetSuccess(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sl := domain.SubscriberLocator{
		ChannelID:    "my-channel",
		SubscriberID: "sbsc-1",
	}
	msgs := make([]domain.Message, 128)
	msgJSONs := make([]interface{}, len(msgs))
	for i := range msgs {
		msgs[i] = domain.Message{
			MessageLocator: domain.MessageLocator{
				ChannelID: sl.ChannelID,
				MessageID: domain.MessageID(fmt.Sprintf("msg-%d", i)),
			},
			Content: json.RawMessage(fmt.Sprintf(`{"hi": "hello %d"}`, i)),
		}
		msgJSONs[i] = map[string]interface{}{
			"messageID": string(msgs[i].MessageID),
			"content": map[string]interface{}{
				"hi": fmt.Sprintf("hello %d", i),
			},
		}
	}
	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {}, func(deps *ServerDependencies, baseURL string) {
		assert.NoError(t, deps.Storage.AsPubSubStorage().NewSubscriber(ctx, sl))

		// No message
		res := DoHTTPRequest(t, "GET", fmt.Sprintf("%s/channel/%s/subscription/polling/%s?timeout=0ms", baseURL, sl.ChannelID, sl.SubscriberID), ``)
		body := AssertResponseJSON(t, res, 200, map[string]interface{}{
			"channelID":    string(sl.ChannelID),
			"messages":     []interface{}{},
			"moreMessages": false,
		})
		assert.NotContains(t, body, "ackHandle")

		// Publish messages
		assert.NoError(t, deps.Storage.AsPubSubStorage().PublishMessages(ctx, msgs))

		// Got messages
		res = DoHTTPRequest(t, "GET", fmt.Sprintf("%s/channel/%s/subscription/polling/%s", baseURL, sl.ChannelID, sl.SubscriberID), ``)
		body = AssertResponseJSON(t, res, 200, map[string]interface{}{
			"channelID":    string(sl.ChannelID),
			"messages":     msgJSONs[:64], // max = default 64
			"moreMessages": true,
		})
		assert.Contains(t, body, "ackHandle")
		res = DoHTTPRequest(t, "GET", fmt.Sprintf("%s/channel/%s/subscription/polling/%s?max=%d", baseURL, sl.ChannelID, sl.SubscriberID, len(msgs)+1), ``)
		body = AssertResponseJSON(t, res, 200, map[string]interface{}{
			"channelID":    string(sl.ChannelID),
			"messages":     msgJSONs, // max = len(msgs) + 1
			"moreMessages": false,
		})
		assert.Contains(t, body, "ackHandle")
	})
}

func TestPollingSubscriberTimeoutAndMax(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	storage, pubsub, _ := NewMockStorages(ctrl)

	sl := domain.SubscriberLocator{
		ChannelID:    "my-channel",
		SubscriberID: "sbsc-1",
	}
	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {
		deps.Storage = storage
	}, func(deps *ServerDependencies, baseURL string) {
		timeout := domain.Duration{Duration: 0 * time.Second}
		max := 16
		pubsub.EXPECT().FetchMessages(gomock.Any(), sl, max, timeout).Return(
			[]domain.Message{},
			false, // moreMsg
			domain.AckHandle{},
			nil,
		)
		res := DoHTTPRequest(t, "GET", fmt.Sprintf("%s/channel/%s/subscription/polling/%s?timeout=%s&max=%d", baseURL, sl.ChannelID, sl.SubscriberID, timeout, max), ``)
		AssertResponseJSON(t, res, 200, map[string]interface{}{
			"channelID":    string(sl.ChannelID),
			"messages":     []interface{}{},
			"moreMessages": false,
		})

		// timeout over configured max
		timeout = domain.Duration{Duration: deps.GetLongPollingMaxTimeout().Duration + (1 * time.Second)}
		max++ // Ensure max value is properly passed to storage backend
		pubsub.EXPECT().FetchMessages(gomock.Any(), sl, max, deps.GetLongPollingMaxTimeout()).Return(
			[]domain.Message{},
			false, // moreMsg
			domain.AckHandle{},
			nil,
		)
		res = DoHTTPRequest(t, "GET", fmt.Sprintf("%s/channel/%s/subscription/polling/%s?timeout=%s&max=%d", baseURL, sl.ChannelID, sl.SubscriberID, timeout, max), ``)
		assert.Equal(t, 200, res.StatusCode)
	})
}

func TestPollingSubscriberGetServerClose(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	storage, pubsub, _ := NewMockStorages(ctrl)

	sl := domain.SubscriberLocator{
		ChannelID:    "my-channel",
		SubscriberID: "sbsc-1",
	}
	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {
		deps.Storage = storage
	}, func(deps *ServerDependencies, baseURL string) {
		deps.ServerClose.Close()

		timeout := domain.Duration{Duration: 10 * time.Second}
		max := 16
		pubsub.EXPECT().FetchMessages(gomock.Any(), sl, max, timeout).DoAndReturn(func(ctx context.Context, _ domain.SubscriberLocator, max int, timeout domain.Duration) ([]domain.Message, bool, domain.AckHandle, error) {
			assert.Error(t, ctx.Err(), "context should be closed")
			return []domain.Message{}, false, domain.AckHandle{}, ctx.Err()
		})
		res := DoHTTPRequest(t, "GET", fmt.Sprintf("%s/channel/%s/subscription/polling/%s?timeout=%s&max=%d", baseURL, sl.ChannelID, sl.SubscriberID, timeout, max), ``)
		AssertResponseJSON(t, res, 200, map[string]interface{}{
			"channelID":    string(sl.ChannelID),
			"messages":     []interface{}{},
			"moreMessages": false,
		})
	})
}

func TestPollingSubscriberGetFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	storage, pubsub, _ := NewMockStorages(ctrl)

	sl := domain.SubscriberLocator{
		ChannelID:    "my-channel",
		SubscriberID: "sbsc-1",
	}
	timeout := domain.Duration{Duration: 10 * time.Second}
	max := 16
	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {
		deps.Storage = storage
	}, func(deps *ServerDependencies, baseURL string) {
		res := DoHTTPRequest(t, "GET", fmt.Sprintf("%s/channel/%s/subscription/polling/%s", baseURL, "*** INVALID ***", sl.SubscriberID), ``)
		AssertErrorResponse(t, res, 400, nil, `Invalid "channelID" parameter`)

		res = DoHTTPRequest(t, "GET", fmt.Sprintf("%s/channel/%s/subscription/polling/%s", baseURL, sl.ChannelID, "*** INVALID ***"), ``)
		AssertErrorResponse(t, res, 400, nil, `Invalid "subscriberID" parameter`)

		res = DoHTTPRequest(t, "GET", fmt.Sprintf("%s/channel/%s/subscription/polling/%s?timeout=INVALID", baseURL, sl.ChannelID, sl.SubscriberID), ``)
		AssertErrorResponse(t, res, 400, nil, `Invalid "timeout" parameter`)

		res = DoHTTPRequest(t, "GET", fmt.Sprintf("%s/channel/%s/subscription/polling/%s?max=INVALID", baseURL, sl.ChannelID, sl.SubscriberID), ``)
		AssertErrorResponse(t, res, 400, nil, `Invalid "max" parameter`)

		pubsub.EXPECT().FetchMessages(gomock.Any(), sl, max, timeout).Return([]domain.Message{}, false, domain.AckHandle{}, domain.ErrInvalidChannel)
		res = DoHTTPRequest(t, "GET", fmt.Sprintf("%s/channel/%s/subscription/polling/%s?timeout=%s&max=%d", baseURL, sl.ChannelID, sl.SubscriberID, timeout, max), ``)
		AssertErrorResponse(t, res, 403, domain.ErrInvalidChannel, "")

		pubsub.EXPECT().FetchMessages(gomock.Any(), sl, max, timeout).Return([]domain.Message{}, false, domain.AckHandle{}, domain.ErrSubscriptionNotFound)
		res = DoHTTPRequest(t, "GET", fmt.Sprintf("%s/channel/%s/subscription/polling/%s?timeout=%s&max=%d", baseURL, sl.ChannelID, sl.SubscriberID, timeout, max), ``)
		AssertErrorResponse(t, res, 404, domain.ErrSubscriptionNotFound, "")

		pubsub.EXPECT().FetchMessages(gomock.Any(), sl, max, timeout).Return([]domain.Message{}, false, domain.AckHandle{}, errors.New("mock error"))
		res = DoHTTPRequest(t, "GET", fmt.Sprintf("%s/channel/%s/subscription/polling/%s?timeout=%s&max=%d", baseURL, sl.ChannelID, sl.SubscriberID, timeout, max), ``)
		AssertInternalServerErrorResponse(t, res)
	})
}

func TestPollingSubscriberMessageDeleteSuccess(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sl := domain.SubscriberLocator{
		ChannelID:    "my-channel",
		SubscriberID: "sbsc-1",
	}
	msgs := make([]domain.Message, 128)
	for i := range msgs {
		msgs[i] = domain.Message{
			MessageLocator: domain.MessageLocator{
				ChannelID: sl.ChannelID,
				MessageID: domain.MessageID(fmt.Sprintf("msg-%d", i)),
			},
			Content: json.RawMessage(fmt.Sprintf(`{"hi": "hello %d"}`, i)),
		}
	}
	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {}, func(deps *ServerDependencies, baseURL string) {
		pubsub := deps.Storage.AsPubSubStorage()
		assert.NoError(t, pubsub.NewSubscriber(ctx, sl))
		assert.NoError(t, pubsub.PublishMessages(ctx, msgs))
		fetched, _, ackHandle, err := pubsub.FetchMessages(ctx, sl, len(msgs)/2, domain.Duration{Duration: 0})
		assert.Equal(t, msgs[:len(msgs)/2], fetched)
		assert.NoError(t, err)

		res := DoHTTPRequest(t, "DELETE", fmt.Sprintf("%s/channel/%s/subscription/polling/%s/message?ackHandle=%s", baseURL, sl.ChannelID, sl.SubscriberID, ackHandle.Handle), ``)
		assert.Equal(t, 204, res.StatusCode)

		fetched, _, _, err = pubsub.FetchMessages(ctx, sl, len(msgs)/2, domain.Duration{Duration: 0})
		assert.Equal(t, msgs[len(msgs)/2:], fetched)
		assert.NoError(t, err)

		// Should be idempotent
		res = DoHTTPRequest(t, "DELETE", fmt.Sprintf("%s/channel/%s/subscription/polling/%s/message?ackHandle=%s", baseURL, sl.ChannelID, sl.SubscriberID, ackHandle.Handle), ``)
		assert.Equal(t, 204, res.StatusCode)
	})
}

func TestPollingSubscriberMessageDeleteFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	storage, pubsub, _ := NewMockStorages(ctrl)

	sl := domain.SubscriberLocator{
		ChannelID:    "my-channel",
		SubscriberID: "sbsc-1",
	}
	ackHandle := domain.AckHandle{SubscriberLocator: sl, Handle: `64852321-C74B-43AC-A893-A6E349F1B476`}
	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {
		deps.Storage = storage
	}, func(deps *ServerDependencies, baseURL string) {
		res := DoHTTPRequest(t, "DELETE", fmt.Sprintf("%s/channel/%s/subscription/polling/%s/message?ackHandle=%s", baseURL, "*** INVALID ***", sl.SubscriberID, ackHandle.Handle), ``)
		AssertErrorResponse(t, res, 400, nil, `Invalid "channelID" parameter`)

		res = DoHTTPRequest(t, "DELETE", fmt.Sprintf("%s/channel/%s/subscription/polling/%s/message?ackHandle=%s", baseURL, sl.ChannelID, "*** INVALID ***", ackHandle.Handle), ``)
		AssertErrorResponse(t, res, 400, nil, `Invalid "subscriberID" parameter`)

		res = DoHTTPRequest(t, "DELETE", fmt.Sprintf("%s/channel/%s/subscription/polling/%s/message?", baseURL, sl.ChannelID, sl.SubscriberID), ``)
		AssertErrorResponse(t, res, 400, nil, `Missing "ackHandle" parameter`)

		pubsub.EXPECT().AcknowledgeMessages(gomock.Any(), ackHandle).Return(domain.ErrInvalidChannel)
		res = DoHTTPRequest(t, "DELETE", fmt.Sprintf("%s/channel/%s/subscription/polling/%s/message?ackHandle=%s", baseURL, sl.ChannelID, sl.SubscriberID, ackHandle.Handle), ``)
		AssertErrorResponse(t, res, 403, domain.ErrInvalidChannel, "")

		pubsub.EXPECT().AcknowledgeMessages(gomock.Any(), ackHandle).Return(domain.ErrMalformedAckHandle)
		res = DoHTTPRequest(t, "DELETE", fmt.Sprintf("%s/channel/%s/subscription/polling/%s/message?ackHandle=%s", baseURL, sl.ChannelID, sl.SubscriberID, ackHandle.Handle), ``)
		AssertErrorResponse(t, res, 400, domain.ErrMalformedAckHandle, "")

		pubsub.EXPECT().AcknowledgeMessages(gomock.Any(), ackHandle).Return(domain.ErrSubscriptionNotFound)
		res = DoHTTPRequest(t, "DELETE", fmt.Sprintf("%s/channel/%s/subscription/polling/%s/message?ackHandle=%s", baseURL, sl.ChannelID, sl.SubscriberID, ackHandle.Handle), ``)
		AssertErrorResponse(t, res, 404, domain.ErrSubscriptionNotFound, "")

		pubsub.EXPECT().AcknowledgeMessages(gomock.Any(), ackHandle).Return(errors.New("mock error"))
		res = DoHTTPRequest(t, "DELETE", fmt.Sprintf("%s/channel/%s/subscription/polling/%s/message?ackHandle=%s", baseURL, sl.ChannelID, sl.SubscriberID, ackHandle.Handle), ``)
		AssertInternalServerErrorResponse(t, res)
	})
}
