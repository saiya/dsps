package endpoints

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/http/lifecycle"
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
func InitSubscriptionPollingEndpoints(router gin.IRouter, deps PollingEndpointDependency) {
	group := router.Group("/subscription/polling/:subscriberID")
	group.Use(func(ctx *gin.Context) {
		logger.ModifyGinContext(ctx).WithStr("subscriberID", ctx.Param("subscriberID")).Build()
		ctx.Next()
	})

	group.PUT("", subscriberPutEndpoint(deps))
	group.DELETE("", subscriberDeleteEndpoint(deps))
	group.GET("", subscriberGetEndpoint(deps))
	group.DELETE("/message", subscriberMessageDeleteEndpoint(deps))
}

func subscriberPutEndpoint(deps PollingEndpointDependency) gin.HandlerFunc {
	pubsub := deps.GetStorage().AsPubSubStorage()
	return func(ctx *gin.Context) {
		if pubsub == nil {
			utils.SendPubSubUnsupportedError(ctx)
			return
		}

		channelID, err := domain.ParseChannelID(ctx.Param("channelID"))
		if err != nil {
			utils.SendInvalidParameter(ctx, "channelID", err)
			return
		}

		subscriberID, err := domain.ParseSubscriberID(ctx.Param("subscriberID"))
		if err != nil {
			utils.SendInvalidParameter(ctx, "subscriberID", err)
			return
		}

		err = pubsub.NewSubscriber(ctx, domain.SubscriberLocator{
			ChannelID:    channelID,
			SubscriberID: subscriberID,
		})
		if err != nil {
			if errors.Is(err, domain.ErrInvalidChannel) {
				// Could not create/access to the channel because not permitted by configuration
				utils.SendError(ctx, http.StatusForbidden, err.Error(), err)
			} else {
				utils.SentInternalServerError(ctx, err)
			}
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"channelID":    channelID,
			"subscriberID": subscriberID,
		})
	}
}

func subscriberDeleteEndpoint(deps PollingEndpointDependency) gin.HandlerFunc {
	pubsub := deps.GetStorage().AsPubSubStorage()
	return func(ctx *gin.Context) {
		if pubsub == nil {
			utils.SendPubSubUnsupportedError(ctx)
			return
		}

		channelID, err := domain.ParseChannelID(ctx.Param("channelID"))
		if err != nil {
			utils.SendInvalidParameter(ctx, "channelID", err)
			return
		}

		subscriberID, err := domain.ParseSubscriberID(ctx.Param("subscriberID"))
		if err != nil {
			utils.SendInvalidParameter(ctx, "subscriberID", err)
			return
		}

		err = pubsub.RemoveSubscriber(ctx, domain.SubscriberLocator{
			ChannelID:    channelID,
			SubscriberID: subscriberID,
		})
		if err != nil {
			if errors.Is(err, domain.ErrInvalidChannel) {
				// Belonging channel might be deleted / expired or not permitted in configuration.
				utils.SendError(ctx, http.StatusForbidden, err.Error(), err)
			} else {
				utils.SentInternalServerError(ctx, err)
			}
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"channelID":    channelID,
			"subscriberID": subscriberID,
		})
	}
}

func subscriberGetEndpoint(deps PollingEndpointDependency) gin.HandlerFunc {
	pubsub := deps.GetStorage().AsPubSubStorage()
	serverClose := deps.GetServerClose()
	longPollingMaxTimeout := deps.GetLongPollingMaxTimeout().Duration
	return func(ctx *gin.Context) {
		if pubsub == nil {
			utils.SendPubSubUnsupportedError(ctx)
			return
		}

		channelID, err := domain.ParseChannelID(ctx.Param("channelID"))
		if err != nil {
			utils.SendInvalidParameter(ctx, "channelID", err)
			return
		}

		subscriberID, err := domain.ParseSubscriberID(ctx.Param("subscriberID"))
		if err != nil {
			utils.SendInvalidParameter(ctx, "subscriberID", err)
			return
		}

		timeout, err := time.ParseDuration(ctx.DefaultQuery("timeout", "0ms"))
		if err != nil {
			utils.SendInvalidParameter(ctx, "timeout", err)
			return
		}
		if timeout > longPollingMaxTimeout {
			logger.Of(ctx).Infof(logger.CatHTTP, "Client requested long-polling timeout %v is too long, rounded to longPollingMaxTimeout (%v)", timeout, longPollingMaxTimeout)
			timeout = longPollingMaxTimeout
		}

		max, err := strconv.ParseInt(ctx.DefaultQuery("max", "64"), 10, 0)
		if err != nil {
			utils.SendInvalidParameter(ctx, "max", err)
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
						utils.SendError(ctx, http.StatusForbidden, err.Error(), err)
					} else if errors.Is(err, domain.ErrSubscriptionNotFound) {
						// Channel / subscriber might be expired or intentionally deleted.
						utils.SendError(ctx, http.StatusNotFound, err.Error(), err)
					} else {
						utils.SentInternalServerError(ctx, err)
					}
					return
				}
			}

			resultMsgs := make([]gin.H, 0, len(msgs))
			for _, msg := range msgs {
				resultMsgs = append(resultMsgs, gin.H{
					"messageID": msg.MessageID,
					"content":   msg.Content,
				})
			}
			result := gin.H{
				"channelID":    channelID,
				"messages":     resultMsgs,
				"moreMessages": moreMsg,
			}
			if len(msgs) > 0 {
				result["ackHandle"] = ackHandle.Handle
			}
			ctx.JSON(http.StatusOK, result)
		})
	}
}

func subscriberMessageDeleteEndpoint(deps PollingEndpointDependency) gin.HandlerFunc {
	pubsub := deps.GetStorage().AsPubSubStorage()
	return func(ctx *gin.Context) {
		if pubsub == nil {
			utils.SendPubSubUnsupportedError(ctx)
			return
		}

		channelID, err := domain.ParseChannelID(ctx.Param("channelID"))
		if err != nil {
			utils.SendInvalidParameter(ctx, "channelID", err)
			return
		}

		subscriberID, err := domain.ParseSubscriberID(ctx.Param("subscriberID"))
		if err != nil {
			utils.SendInvalidParameter(ctx, "subscriberID", err)
			return
		}

		ackHandle := ctx.Query("ackHandle")
		if ackHandle == "" {
			utils.SendMissingParameter(ctx, "ackHandle")
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
				utils.SendError(ctx, http.StatusForbidden, err.Error(), err)
			} else if errors.Is(err, domain.ErrMalformedAckHandle) {
				utils.SendError(ctx, http.StatusBadRequest, err.Error(), err)
			} else if errors.Is(err, domain.ErrSubscriptionNotFound) {
				// Belonging channel/subscriber could be expired/deleted.
				utils.SendError(ctx, http.StatusNotFound, err.Error(), err)
			} else {
				utils.SentInternalServerError(ctx, err)
			}
			return
		}

		ctx.Status(http.StatusNoContent)
	}
}
