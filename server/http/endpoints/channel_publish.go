package endpoints

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/http/router"
	"github.com/saiya/dsps/server/http/utils"
)

// PublishEndpointDependency is to inject required objects to the endpoint
type PublishEndpointDependency interface {
	GetStorage() domain.Storage
}

// InitPublishEndpoints registers endpoints
func InitPublishEndpoints(rt *router.Router, deps PublishEndpointDependency) {
	pubsub := deps.GetStorage().AsPubSubStorage()

	rt.PUT("/message/:messageID", func(ctx context.Context, args router.HandlerArgs) {
		if pubsub == nil {
			utils.SendPubSubUnsupportedError(ctx, args.W)
			return
		}

		channelID, err := domain.ParseChannelID(args.PS.ByName("channelID"))
		if err != nil {
			utils.SendInvalidParameter(ctx, args.W, "channelID", err)
			return
		}

		messageID, err := domain.ParseMessageID(args.PS.ByName("messageID"))
		if err != nil {
			utils.SendInvalidParameter(ctx, args.W, "messageID", err)
			return
		}

		content, err := args.R.ReadBody()
		if err == nil && !json.Valid(content) {
			err = xerrors.New("Is not valid JSON")
		}
		if err != nil {
			utils.SendError(ctx, args.W, http.StatusBadRequest, "Request body is not JSON", err)
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
				utils.SendError(ctx, args.W, http.StatusForbidden, err.Error(), err)
			} else {
				utils.SendInternalServerError(ctx, args.W, err)
			}
			return
		}

		utils.SendJSON(ctx, args.W, http.StatusOK, map[string]interface{}{
			"channelID": channelID,
			"messageID": messageID,
		})
	})
}
