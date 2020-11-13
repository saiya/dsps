package endpoints

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/saiya/dsps/server/domain"
)

func sendError(ctx *gin.Context, status int, message string, err error) {
	// TODO: Output error to log
	fmt.Printf("sendError(status = %d, message = \"%s\", err = %#v)\n", status, message, err)

	var code *string = nil
	if errWithCode := domain.NewErrorWithCode(""); errors.As(err, &errWithCode) {
		code = errWithCode.Code()
	}

	ctx.AbortWithStatusJSON(status, gin.H{
		"error":  message,
		"detail": err, // TODO: Hide err detail in production mode
		"code":   code,
	})
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
