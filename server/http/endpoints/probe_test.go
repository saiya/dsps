package endpoints_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/domain/mock"
	. "github.com/saiya/dsps/server/http"
	. "github.com/saiya/dsps/server/http/testing"
)

func TestProbeSuccess(t *testing.T) {
	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {}, func(deps *ServerDependencies, baseURL string) {
		res := DoHTTPRequest(t, "GET", baseURL+"/probe/liveness", "")
		assert.Equal(t, 200, res.StatusCode)

		res = DoHTTPRequest(t, "GET", baseURL+"/probe/readiness", "")
		assert.Equal(t, 200, res.StatusCode)
	})
}

func TestProbeFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	storage := NewMockStorage(ctrl)
	storage.EXPECT().AsPubSubStorage().Return(nil).AnyTimes()
	storage.EXPECT().AsJwtStorage().Return(nil).AnyTimes()

	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {
		deps.Storage = storage
	}, func(deps *ServerDependencies, baseURL string) {
		storage.EXPECT().Liveness(gomock.Any()).Return(nil, errors.New("mock error"))
		res := DoHTTPRequest(t, "GET", baseURL+"/probe/liveness", "")
		assert.Equal(t, 500, res.StatusCode)

		storage.EXPECT().Readiness(gomock.Any()).Return(nil, errors.New("mock error"))
		res = DoHTTPRequest(t, "GET", baseURL+"/probe/readiness", "")
		assert.Equal(t, 500, res.StatusCode)
	})
}
