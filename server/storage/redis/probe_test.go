package redis

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	dspstesting "github.com/saiya/dsps/server/testing"
	"github.com/stretchr/testify/assert"
)

func TestProbeErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	errToRetrun := errors.New("Mocked redis error")

	s, redisCmd := newMockedRedisStorage(ctrl)
	redisCmd.EXPECT().Ping(gomock.Any()).Return(errToRetrun)

	_, err := s.Readiness(context.Background())
	dspstesting.IsError(t, errToRetrun, err)

	_, err = s.Liveness(context.Background())
	assert.NoError(t, err) // Always success
}
