package endpoints

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/http/lifecycle"
	"github.com/saiya/dsps/server/http/router"
	"github.com/saiya/dsps/server/http/utils"
	"github.com/saiya/dsps/server/logger"
)

// PollingEndpointDependency is to inject required objects to the endpoint
type PollingEndpointDependency interface {
	GetServerClose() lifecycle.ServerClose
	GetStorage() domain.Storage

	GetLongPollingMaxTimeout() domain.Duration
}

// InitSubscriptionPollingEndpoints registers endpoints
func InitSubscriptionPollingEndpoints(channelRouter *router.Router, deps PollingEndpointDependency) {
	group := channelRouter.NewGroup(
		"/subscription/polling/:subscriberID",
		func(ctx context.Context, args router.MiddlewareArgs, next func(context.Context)) {
			next(logger.WithAttributes(ctx).WithStr("subscriberID", args.PS.ByName("subscriberID")).Build())
		},
	)
	group.PUT("", subscriberPutEndpoint(deps))
	group.DELETE("", subscriberDeleteEndpoint(deps))
	group.GET("", subscriberGetEndpoint(deps))
	group.DELETE("/message", subscriberMessageDeleteEndpoint(deps))
}

func subscriberPutEndpoint(deps PollingEndpointDependency) router.Handler {
	pubsub := deps.GetStorage().AsPubSubStorage()
	return func(ctx context.Context, args router.HandlerArgs) {
		if pubsub == nil {
			utils.SendPubSubUnsupportedError(ctx, args.W)
			return
		}

		channelID, err := domain.ParseChannelID(args.PS.ByName("channelID"))
		if err != nil {
			utils.SendInvalidParameter(ctx, args.W, "channelID", err)
			return
		}

		subscriberID, err := domain.ParseSubscriberID(args.PS.ByName("subscriberID"))
		if err != nil {
			utils.SendInvalidParameter(ctx, args.W, "subscriberID", err)
			return
		}

		err = pubsub.NewSubscriber(ctx, domain.SubscriberLocator{
			ChannelID:    channelID,
			SubscriberID: subscriberID,
		})
		if err != nil {
			if errors.Is(err, domain.ErrInvalidChannel) {
				// Could not create/access to the channel because not permitted by configuration
				utils.SendError(ctx, args.W, http.StatusForbidden, err.Error(), err)
			} else {
				utils.SendInternalServerError(ctx, args.W, err)
			}
			return
		}

		utils.SendJSON(ctx, args.W, http.StatusOK, map[string]interface{}{
			"channelID":    channelID,
			"subscriberID": subscriberID,
		})
	}
}

func subscriberDeleteEndpoint(deps PollingEndpointDependency) router.Handler {
	pubsub := deps.GetStorage().AsPubSubStorage()
	return func(ctx context.Context, args router.HandlerArgs) {
		if pubsub == nil {
			utils.SendPubSubUnsupportedError(ctx, args.W)
			return
		}

		channelID, err := domain.ParseChannelID(args.PS.ByName("channelID"))
		if err != nil {
			utils.SendInvalidParameter(ctx, args.W, "channelID", err)
			return
		}

		subscriberID, err := domain.ParseSubscriberID(args.PS.ByName("subscriberID"))
		if err != nil {
			utils.SendInvalidParameter(ctx, args.W, "subscriberID", err)
			return
		}

		err = pubsub.RemoveSubscriber(ctx, domain.SubscriberLocator{
			ChannelID:    channelID,
			SubscriberID: subscriberID,
		})
		if err != nil {
			if errors.Is(err, domain.ErrInvalidChannel) {
				// Belonging channel might be deleted / expired or not permitted in configuration.
				utils.SendError(ctx, args.W, http.StatusForbidden, err.Error(), err)
			} else {
				utils.SendInternalServerError(ctx, args.W, err)
			}
			return
		}

		utils.SendJSON(ctx, args.W, http.StatusOK, map[string]interface{}{
			"channelID":    channelID,
			"subscriberID": subscriberID,
		})
	}
}

func subscriberGetEndpoint(deps PollingEndpointDependency) router.Handler {
	pubsub := deps.GetStorage().AsPubSubStorage()
	serverClose := deps.GetServerClose()
	longPollingMaxTimeout := deps.GetLongPollingMaxTimeout().Duration
	return func(ctx context.Context, args router.HandlerArgs) {
		if pubsub == nil {
			utils.SendPubSubUnsupportedError(ctx, args.W)
			return
		}

		channelID, err := domain.ParseChannelID(args.PS.ByName("channelID"))
		if err != nil {
			utils.SendInvalidParameter(ctx, args.W, "channelID", err)
			return
		}

		subscriberID, err := domain.ParseSubscriberID(args.PS.ByName("subscriberID"))
		if err != nil {
			utils.SendInvalidParameter(ctx, args.W, "subscriberID", err)
			return
		}

		timeout, err := time.ParseDuration(args.R.GetQueryParamOrDefault("timeout", "0ms"))
		if err != nil {
			utils.SendInvalidParameter(ctx, args.W, "timeout", err)
			return
		}
		if timeout > longPollingMaxTimeout {
			logger.Of(ctx).Infof(logger.CatHTTP, "Client requested long-polling timeout %v is too long, rounded to longPollingMaxTimeout (%v)", timeout, longPollingMaxTimeout)
			timeout = longPollingMaxTimeout
		}

		max, err := strconv.ParseInt(args.R.GetQueryParamOrDefault("max", "64"), 10, 0)
		if err != nil {
			utils.SendInvalidParameter(ctx, args.W, "max", err)
			return
		}

		serverClose.WithCancel(ctx, func(ctxWithCancel context.Context) {
			msgs, moreMsg, ackHandle, err := pubsub.FetchMessages(
				ctxWithCancel, // Stop polling on server close.
				domain.SubscriberLocator{
					ChannelID:    channelID,
					SubscriberID: subscriberID,
				},
				int(max),
				domain.Duration{Duration: timeout},
			)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					logger.Of(ctx).Infof(logger.CatHTTP, "Polling canceled due to context cancel, returned empty messages to client.")
					msgs = []domain.Message{}
					moreMsg = false
					ackHandle = domain.AckHandle{}
					// Continue to normal flow
				} else {
					if errors.Is(err, domain.ErrInvalidChannel) {
						utils.SendError(ctx, args.W, http.StatusForbidden, err.Error(), err)
					} else if errors.Is(err, domain.ErrSubscriptionNotFound) {
						// Channel / subscriber might be expired or intentionally deleted.
						utils.SendError(ctx, args.W, http.StatusNotFound, err.Error(), err)
					} else {
						utils.SendInternalServerError(ctx, args.W, err)
					}
					return
				}
			}

			resultMsgs := make([]interface{}, 0, len(msgs))
			for _, msg := range msgs {
				resultMsgs = append(resultMsgs, map[string]interface{}{
					"messageID": msg.MessageID,
					"content":   msg.Content,
				})
			}
			result := map[string]interface{}{
				"channelID":    channelID,
				"messages":     resultMsgs,
				"moreMessages": moreMsg,
			}
			if len(msgs) > 0 {
				result["ackHandle"] = ackHandle.Handle
			}
			utils.SendJSON(ctx, args.W, 200, result)
		})
	}
}

func subscriberMessageDeleteEndpoint(deps PollingEndpointDependency) router.Handler {
	pubsub := deps.GetStorage().AsPubSubStorage()
	return func(ctx context.Context, args router.HandlerArgs) {
		if pubsub == nil {
			utils.SendPubSubUnsupportedError(ctx, args.W)
			return
		}

		channelID, err := domain.ParseChannelID(args.PS.ByName("channelID"))
		if err != nil {
			utils.SendInvalidParameter(ctx, args.W, "channelID", err)
			return
		}

		subscriberID, err := domain.ParseSubscriberID(args.PS.ByName("subscriberID"))
		if err != nil {
			utils.SendInvalidParameter(ctx, args.W, "subscriberID", err)
			return
		}

		ackHandle := args.R.GetQueryParam("ackHandle")
		if ackHandle == "" {
			utils.SendMissingParameter(ctx, args.W, "ackHandle")
			return
		}

		err = pubsub.AcknowledgeMessages(ctx, domain.AckHandle{
			SubscriberLocator: domain.SubscriberLocator{
				ChannelID:    channelID,
				SubscriberID: subscriberID,
			},
			Handle: ackHandle,
		})
		if err != nil {
			if errors.Is(err, domain.ErrInvalidChannel) {
				utils.SendError(ctx, args.W, http.StatusForbidden, err.Error(), err)
			} else if errors.Is(err, domain.ErrMalformedAckHandle) {
				utils.SendError(ctx, args.W, http.StatusBadRequest, err.Error(), err)
			} else if errors.Is(err, domain.ErrSubscriptionNotFound) {
				// Belonging channel/subscriber could be expired/deleted.
				utils.SendError(ctx, args.W, http.StatusNotFound, err.Error(), err)
			} else {
				utils.SendInternalServerError(ctx, args.W, err)
			}
			return
		}

		utils.SendNoContent(ctx, args.W)
	}
}
