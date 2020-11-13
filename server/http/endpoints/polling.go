package endpoints

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/saiya/dsps/server/domain"
)

// PollingEndpointDependency is to inject required objects to the endpoint
type PollingEndpointDependency interface {
	GetStorage() domain.Storage
}

// InitPollingEndpoints registers endpoints
func InitPollingEndpoints(router gin.IRoutes, deps ProbeEndpointDependency) {
	pubsub := deps.GetStorage().AsPubSubStorage()

	router.PUT("/channel/:channelID/subscription/polling/:subscriberID", func(ctx *gin.Context) {
		if pubsub == nil {
			sendPubSubUnsupportedError(ctx)
			return
		}

		channelID, err := domain.ParseChannelID(ctx.Param("channelID"))
		if err != nil {
			sendInvalidParameter(ctx, "channelID", err)
			return
		}

		subscriberID, err := domain.ParseSubscriberID(ctx.Param("subscriberID"))
		if err != nil {
			sendInvalidParameter(ctx, "subscriberID", err)
			return
		}

		err = pubsub.NewSubscriber(ctx, domain.SubscriberLocator{
			ChannelID:    channelID,
			SubscriberID: subscriberID,
		})
		if err != nil {
			if errors.Is(err, domain.ErrInvalidChannel) {
				// Could not create/access to the channel because not permitted by configuration
				sendError(ctx, http.StatusForbidden, err.Error(), err)
			} else {
				sentInternalServerError(ctx, err)
			}
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"channelID":    channelID,
			"subscriberID": subscriberID,
		})
	})

	router.DELETE("/channel/:channelID/subscription/polling/:subscriberID", func(ctx *gin.Context) {
		if pubsub == nil {
			sendPubSubUnsupportedError(ctx)
			return
		}

		channelID, err := domain.ParseChannelID(ctx.Param("channelID"))
		if err != nil {
			sendInvalidParameter(ctx, "channelID", err)
			return
		}

		subscriberID, err := domain.ParseSubscriberID(ctx.Param("subscriberID"))
		if err != nil {
			sendInvalidParameter(ctx, "subscriberID", err)
			return
		}

		err = pubsub.RemoveSubscriber(ctx, domain.SubscriberLocator{
			ChannelID:    channelID,
			SubscriberID: subscriberID,
		})
		if err != nil {
			if errors.Is(err, domain.ErrInvalidChannel) {
				// Belonging channel might be deleted / expired or not permitted in configuration.
				sendError(ctx, http.StatusForbidden, err.Error(), err)
			} else {
				sentInternalServerError(ctx, err)
			}
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"channelID":    channelID,
			"subscriberID": subscriberID,
		})
	})

	router.GET("/channel/:channelID/subscription/polling/:subscriberID", func(ctx *gin.Context) {
		if pubsub == nil {
			sendPubSubUnsupportedError(ctx)
			return
		}

		channelID, err := domain.ParseChannelID(ctx.Param("channelID"))
		if err != nil {
			sendInvalidParameter(ctx, "channelID", err)
			return
		}

		subscriberID, err := domain.ParseSubscriberID(ctx.Param("subscriberID"))
		if err != nil {
			sendInvalidParameter(ctx, "subscriberID", err)
			return
		}

		timeout, err := time.ParseDuration(ctx.DefaultQuery("timeout", "0ms"))
		if err != nil {
			sendInvalidParameter(ctx, "timeout", err)
			return
		}

		max, err := strconv.ParseInt(ctx.DefaultQuery("max", "64"), 10, 0)
		if err != nil {
			sendInvalidParameter(ctx, "max", err)
			return
		}

		msgs, moreMsg, ackHandle, err := pubsub.FetchMessages(
			ctx,
			domain.SubscriberLocator{
				ChannelID:    channelID,
				SubscriberID: subscriberID,
			},
			int(max),
			domain.Duration{Duration: timeout},
		)
		if err != nil {
			if errors.Is(err, domain.ErrInvalidChannel) {
				sendError(ctx, http.StatusForbidden, err.Error(), err)
			} else if errors.Is(err, domain.ErrSubscriptionNotFound) {
				// Channel / subscriber might be expired or intentionally deleted.
				sendError(ctx, http.StatusNotFound, err.Error(), err)
			} else {
				sentInternalServerError(ctx, err)
			}
			return
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

	router.DELETE("/channel/:channelID/subscription/polling/:subscriberID/message", func(ctx *gin.Context) {
		if pubsub == nil {
			sendPubSubUnsupportedError(ctx)
			return
		}

		channelID, err := domain.ParseChannelID(ctx.Param("channelID"))
		if err != nil {
			sendInvalidParameter(ctx, "channelID", err)
			return
		}

		subscriberID, err := domain.ParseSubscriberID(ctx.Param("subscriberID"))
		if err != nil {
			sendInvalidParameter(ctx, "subscriberID", err)
			return
		}

		ackHandle := ctx.Query("ackHandle")
		if ackHandle == "" {
			sendMissingParameter(ctx, "ackHandle")
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
				sendError(ctx, http.StatusForbidden, err.Error(), err)
			} else if errors.Is(err, domain.ErrMalformedAckHandle) {
				sendError(ctx, http.StatusBadRequest, err.Error(), err)
			} else if errors.Is(err, domain.ErrSubscriptionNotFound) {
				// Belonging channel/subscriber could be expired/deleted.
				sendError(ctx, http.StatusNotFound, err.Error(), err)
			} else {
				sentInternalServerError(ctx, err)
			}
			return
		}

		ctx.Status(http.StatusNoContent)
	})
}
