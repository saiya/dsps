package testing

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/domain"
)

// JwtTest tests common Storage behaviors
func JwtTest(t *testing.T, storageCtor StorageCtor) {
	storageSubTest(t, storageCtor, "JWTScenario", _jwtScenarioTest)
	storageSubTest(t, storageCtor, "JWTPastExp", _jwtPastExpTest)
}

func _jwtScenarioTest(t *testing.T, storageCtor StorageCtor) {
	ctx := context.Background()
	s, err := storageCtor(ctx, domain.RealSystemClock, StubChannelProvider)
	if !assert.NoError(t, err) {
		return
	}
	defer func() { assert.NoError(t, s.Shutdown(ctx)) }()
	storage := s.AsJwtStorage()
	assert.NotNil(t, storage)

	jti := _randomJti()

	result, err := storage.IsRevokedJwt(ctx, jti)
	assert.NoError(t, err)
	assert.False(t, result)

	assert.Nil(t, storage.RevokeJwt(ctx, domain.JwtExp(time.Now().Add(24*time.Hour)), jti))

	result, err = storage.IsRevokedJwt(ctx, jti)
	assert.NoError(t, err)
	assert.True(t, result)
}

func _jwtPastExpTest(t *testing.T, storageCtor StorageCtor) {
	ctx := context.Background()
	s, err := storageCtor(ctx, domain.RealSystemClock, StubChannelProvider)
	if !assert.NoError(t, err) {
		return
	}
	defer func() { assert.NoError(t, s.Shutdown(ctx)) }()
	storage := s.AsJwtStorage()
	assert.NotNil(t, storage)

	jti := _randomJti()

	assert.Nil(t, storage.RevokeJwt(ctx, domain.JwtExp(time.Now().Add(-10*time.Minute)), jti))
	result, err := storage.IsRevokedJwt(ctx, jti)
	assert.NoError(t, err)
	assert.False(t, result)
}

func _randomJti() domain.JwtJti {
	uuid, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	return domain.JwtJti(fmt.Sprintf("jti-%s", uuid))
}
