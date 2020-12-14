package endpoints_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/http"
	. "github.com/saiya/dsps/server/http/testing"
	"github.com/saiya/dsps/server/logger"
)

func TestLogLevelChangeSuccess(t *testing.T) {
	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {}, func(deps *ServerDependencies, baseURL string) {
		assert.False(t, deps.LogFilter.Filter(logger.DEBUG, logger.CatLogger))

		res := DoHTTPRequest(t, "PUT", baseURL+"/admin/log/level?category=logger&level=DEBUG", "")
		assert.Equal(t, 204, res.StatusCode)
		assert.True(t, deps.LogFilter.Filter(logger.DEBUG, logger.CatLogger))
	})
}

func TestLogLevelChangeFailure(t *testing.T) {
	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {}, func(deps *ServerDependencies, baseURL string) {
		// category parameter error
		AssertErrorResponse(
			t,
			DoHTTPRequest(t, "PUT", baseURL+"/admin/log/level?level=DEBUG", ""),
			400,
			nil,
			`Missing "category" parameter`,
		)
		AssertErrorResponse(
			t,
			DoHTTPRequest(t, "PUT", baseURL+"/admin/log/level?category=&level=DEBUG", ""),
			400,
			nil,
			`Missing "category" parameter`,
		)

		// level parameter error
		AssertErrorResponse(
			t,
			DoHTTPRequest(t, "PUT", baseURL+"/admin/log/level?category=logger", ""),
			400,
			nil,
			`Invalid "level" parameter`,
		)
		AssertErrorResponse(
			t,
			DoHTTPRequest(t, "PUT", baseURL+"/admin/log/level?category=logger&level=INVALID", ""),
			400,
			nil,
			`Invalid "level" parameter`,
		)
	})
}
