package endpoints

import (
	"encoding/json"
	"errors"
	"net/http"

	"golang.org/x/xerrors"

	"github.com/gin-gonic/gin"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/http/utils"
)

// PublishEndpointDependency is to inject required objects to the endpoint
type PublishEndpointDependency interface {
	GetStorage() domain.Storage
}

// InitPublishEndpoints registers endpoints
func InitPublishEndpoints(router gin.IRoutes, deps PublishEndpointDependency) {
	pubsub := deps.GetStorage().AsPubSubStorage()

	router.PUT("/message/:messageID", func(ctx *gin.Context) {
		if pubsub == nil {
			utils.SendPubSubUnsupportedError(ctx)
			return
		}

		channelID, err := domain.ParseChannelID(ctx.Param("channelID"))
		if err != nil {
			utils.SendInvalidParameter(ctx, "channelID", err)
			return
		}

		messageID, err := domain.ParseMessageID(ctx.Param("messageID"))
		if err != nil {
			utils.SendInvalidParameter(ctx, "messageID", err)
			return
		}

		content, err := ctx.GetRawData()
		if err == nil && !json.Valid(content) {
			err = xerrors.New("Is not valid JSON")
		}
		if err != nil {
			utils.SendError(ctx, http.StatusBadRequest, "Could not get request body", err)
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
				utils.SendError(ctx, http.StatusForbidden, err.Error(), err)
			} else {
				utils.SentInternalServerError(ctx, err)
			}
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"channelID": channelID,
			"messageID": messageID,
		})
	})
}
