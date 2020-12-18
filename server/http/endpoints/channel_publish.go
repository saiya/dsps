package endpoints

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/http/router"
	"github.com/saiya/dsps/server/http/utils"
	"github.com/saiya/dsps/server/logger"
)

// PublishEndpointDependency is to inject required objects to the endpoint
type PublishEndpointDependency interface {
	GetStorage() domain.Storage
	GetChannelProvider() domain.ChannelProvider
}

// InitPublishEndpoints registers endpoints
func InitPublishEndpoints(channelRouter *router.Router, deps PublishEndpointDependency) {
	pubsub := deps.GetStorage().AsPubSubStorage()

	channelRouter.PUT("/message/:messageID", func(ctx context.Context, args router.HandlerArgs) {
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

		message := domain.Message{
			MessageLocator: domain.MessageLocator{
				ChannelID: channelID,
				MessageID: messageID,
			},
			Content: content,
		}

		err = pubsub.PublishMessages(ctx, []domain.Message{message})
		if err != nil {
			if errors.Is(err, domain.ErrInvalidChannel) {
				// Could not create/access to the channel because not permitted by configuration
				utils.SendError(ctx, args.W, http.StatusForbidden, err.Error(), err)
			} else {
				utils.SendInternalServerError(ctx, args.W, err)
			}
			return
		}

		ch, err := deps.GetChannelProvider().Get(channelID)
		if err != nil {
			utils.SendInternalServerError(ctx, args.W, err)
			return
		}
		if err := ch.SendOutgoingWebhook(ctx, message); err != nil {
			logger.Of(ctx).WarnError(logger.CatOutgoingWebhook, fmt.Sprintf(`failed to send outgoing-webhook (channel: %s, msgID: %s): %%w`, channelID, messageID), err)
		}

		utils.SendJSON(ctx, args.W, http.StatusOK, map[string]interface{}{
			"channelID": channelID,
			"messageID": messageID,
		})
	})
}
