package utils

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/logger"
)

// SendError send arbital error response with given HTTP status.
func SendError(ctx *gin.Context, status int, message string, err error) {
	res := gin.H{"error": message}
	if errWithCode := domain.NewErrorWithCode(""); errors.As(err, &errWithCode) {
		res["code"] = errWithCode.Code()
		logger.ModifyGinContext(ctx).WithStr("code", errWithCode.Code()).Build()
	}

	logger.Of(ctx).InfoError(logger.CatHTTP, fmt.Sprintf("Sending error to client: %s", message), err)
	ctx.AbortWithStatusJSON(status, res)
}

// SentInternalServerError send 500
func SentInternalServerError(ctx *gin.Context, err error) {
	SendError(ctx, http.StatusInternalServerError, "Internal Server Error", err)
}

// SendMissingParameter send 400
func SendMissingParameter(ctx *gin.Context, name string) {
	SendError(
		ctx,
		http.StatusBadRequest,
		fmt.Sprintf("Missing \"%s\" parameter", name),
		nil,
	)
}

// SendInvalidParameter send 400
func SendInvalidParameter(ctx *gin.Context, name string, err error) {
	SendError(
		ctx,
		http.StatusBadRequest,
		fmt.Sprintf("Invalid \"%s\" parameter", name),
		err,
	)
}

// SendPubSubUnsupportedError send 501
func SendPubSubUnsupportedError(ctx *gin.Context) {
	SendError(ctx, http.StatusNotImplemented, "No PubSub compatible storage available.", nil)
}

// SendJwtUnsupportedError send 501
func SendJwtUnsupportedError(ctx *gin.Context) {
	SendError(ctx, http.StatusNotImplemented, "No JWT compatible storage available.", nil)
}
