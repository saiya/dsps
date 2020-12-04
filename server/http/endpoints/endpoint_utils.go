package endpoints

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/logger"
)

func sendError(ctx *gin.Context, status int, message string, err error) {
	res := gin.H{"error": message}
	if errWithCode := domain.NewErrorWithCode(""); errors.As(err, &errWithCode) {
		res["code"] = errWithCode.Code()
		logger.ModifyGinContext(ctx).WithStr("code", errWithCode.Code()).Build()
	}

	logger.Of(ctx).InfoError(fmt.Sprintf("Sending error to client: %s", message), err)
	ctx.AbortWithStatusJSON(status, res)
}

func sentInternalServerError(ctx *gin.Context, err error) {
	sendError(ctx, http.StatusInternalServerError, "Internal Server Error", err)
}

func sendMissingParameter(ctx *gin.Context, name string) {
	sendError(
		ctx,
		http.StatusBadRequest,
		fmt.Sprintf("Missing \"%s\" parameter", name),
		nil,
	)
}

func sendInvalidParameter(ctx *gin.Context, name string, err error) {
	sendError(
		ctx,
		http.StatusBadRequest,
		fmt.Sprintf("Invalid \"%s\" parameter", name),
		err,
	)
}

func sendPubSubUnsupportedError(ctx *gin.Context) {
	sendError(ctx, http.StatusNotImplemented, "No PubSub compatible storage available.", nil)
}
