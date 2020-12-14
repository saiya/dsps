package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/logger"
)

// SendNoContent sends 204 No Content
func SendNoContent(ctx context.Context, w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// SendJSON send JSON response body
func SendJSON(ctx context.Context, w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		panic(xerrors.Errorf("failed to write response body: %w", err))
	}
}

// SendError send arbital error response with given HTTP status.
func SendError(ctx context.Context, w http.ResponseWriter, status int, message string, err error) {
	res := map[string]interface{}{"error": message}
	if errWithCode := domain.NewErrorWithCode(""); errors.As(err, &errWithCode) {
		res["code"] = errWithCode.Code()
		ctx = logger.WithAttributes(ctx).WithStr("code", errWithCode.Code()).Build()
	}

	logger.Of(ctx).InfoError(logger.CatHTTP, fmt.Sprintf("Sending error to client: %s", message), err)
	SendJSON(ctx, w, status, res)
}

// SendInternalServerError send 500
func SendInternalServerError(ctx context.Context, w http.ResponseWriter, err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Of(ctx).Debugf(logger.CatHTTP, "could not send 500 response (may be response already sent): %v", r)
		}
	}()

	logger.Of(ctx).Error("internal server error caught on HTTP endpoint", err)
	w.WriteHeader(500) // Do not send response body. It could be appended to response body sent before.
}

// SendMissingParameter send 400
func SendMissingParameter(ctx context.Context, w http.ResponseWriter, name string) {
	SendError(
		ctx,
		w,
		http.StatusBadRequest,
		fmt.Sprintf("Missing \"%s\" parameter", name),
		nil,
	)
}

// SendInvalidParameter send 400
func SendInvalidParameter(ctx context.Context, w http.ResponseWriter, name string, err error) {
	SendError(
		ctx,
		w,
		http.StatusBadRequest,
		fmt.Sprintf("Invalid \"%s\" parameter", name),
		err,
	)
}

// SendPubSubUnsupportedError send 501
func SendPubSubUnsupportedError(ctx context.Context, w http.ResponseWriter) {
	SendError(ctx, w, http.StatusNotImplemented, "No PubSub compatible storage available.", nil)
}

// SendJwtUnsupportedError send 501
func SendJwtUnsupportedError(ctx context.Context, w http.ResponseWriter) {
	SendError(ctx, w, http.StatusNotImplemented, "No JWT compatible storage available.", nil)
}
