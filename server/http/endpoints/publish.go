package endpoints

import (
	"encoding/json"
	"errors"
	"net/http"

	"golang.org/x/xerrors"

	"github.com/gin-gonic/gin"

	"github.com/saiya/dsps/server/domain"
)

// PublishEndpointDependency is to inject required objects to the endpoint
type PublishEndpointDependency interface {
	GetStorage() domain.Storage
}

// InitPublishEndpoints registers endpoints
func InitPublishEndpoints(router gin.IRoutes, deps ProbeEndpointDependency) {
	pubsub := deps.GetStorage().AsPubSubStorage()

	router.PUT("/channel/:channelID/message/:messageID", func(ctx *gin.Context) {
		if pubsub == nil {
			sendPubSubUnsupportedError(ctx)
			return
		}

		channelID, err := domain.ParseChannelID(ctx.Param("channelID"))
		if err != nil {
			sendInvalidParameter(ctx, "channelID", err)
			return
		}

		messageID, err := domain.ParseMessageID(ctx.Param("messageID"))
		if err != nil {
			sendInvalidParameter(ctx, "messageID", err)
			return
		}

		content, err := ctx.GetRawData()
		if err == nil && !json.Valid(content) {
			err = xerrors.New("Is not valid JSON")
		}
		if err != nil {
			sendError(ctx, http.StatusBadRequest, "Could not get request body", err)
			return
		}

		err = pubsub.PublishMessages(ctx, []domain.Message{
			{
				MessageLocator: domain.MessageLocator{
					ChannelID: channelID,
					MessageID: messageID,
				},
				Content: content,
			},
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
			"channelID": channelID,
			"messageID": messageID,
		})
	})
}
