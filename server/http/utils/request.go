package utils

import (
	"context"
	"regexp"

	"github.com/saiya/dsps/server/http/router"
)

var bearerTokenRegex = regexp.MustCompile("^Bearer ")

// GetBearerToken returns "Authorization: Bearer ..." value.
// If not found, returns "".
func GetBearerToken(ctx context.Context, args router.MiddlewareArgs) string {
	parts := bearerTokenRegex.Split(args.R.Header.Get("Authorization"), 2)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}
