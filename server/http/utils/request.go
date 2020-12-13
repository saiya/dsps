package utils

import (
	"regexp"

	"github.com/gin-gonic/gin"
)

var bearerTokenRegex = regexp.MustCompile("^Bearer ")

// GetBearerToken returns "Authorization: Bearer ..." value.
// If not found, returns "".
func GetBearerToken(ctx *gin.Context) string {
	parts := bearerTokenRegex.Split(ctx.GetHeader("Authorization"), 2)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}
