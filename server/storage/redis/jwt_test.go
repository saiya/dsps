package redis

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/saiya/dsps/server/domain"
	dspstesting "github.com/saiya/dsps/server/testing"
	"github.com/stretchr/testify/assert"
)

func TestJwtRedisErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jti := domain.JwtJti("jwt-id")
	errToReturn := errors.New("Mocked redis error")

	s, redisCmd := newMockedRedisStorage(ctrl)
	redisCmd.EXPECT().Get(gomock.Any(), keyOfJti(jti).Revocation()).Return(nil, errToReturn)

	result, err := s.IsRevokedJwt(context.Background(), jti)
	dspstesting.IsError(t, errToReturn, err)
	assert.Equal(t, false, result)
}
