package testing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/domain"
	dspstesting "github.com/saiya/dsps/server/testing"
)

// JwtTest tests common Storage behaviors
func JwtTest(t *testing.T, storageCtor StorageCtor) {
	ctx := context.Background()
	s, err := storageCtor(ctx, domain.RealSystemClock, StubChannelProvider)
	if !assert.NoError(t, err) {
		return
	}
	defer func() { assert.NoError(t, s.Shutdown(ctx)) }()
	storage := s.AsJwtStorage()
	assert.NotNil(t, storage)

	result, err := storage.IsRevokedJwt(ctx, "my-jwt")
	assert.NoError(t, err)
	assert.False(t, result)

	assert.Nil(t, storage.RevokeJwt(ctx, domain.JwtExp(time.Now().Add(dspstesting.MakeDuration("8760h").Duration)), "my-jwt"))

	result, err = storage.IsRevokedJwt(ctx, "my-jwt")
	assert.NoError(t, err)
	assert.True(t, result)
}
