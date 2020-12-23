package redis

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	storagetesting "github.com/saiya/dsps/server/storage/testing"
	"github.com/saiya/dsps/server/telemetry"
)

func TestInitialConnectFailure(t *testing.T) {
	cfg, err := config.ParseConfig(context.Background(), config.Overrides{}, fmt.Sprintf(`storages: { myRedis: { redis: { singleNode: "%s", timeout: { connect: 1ms }, connection: { max: 10 } } } }`, "127.0.0.1:9999"))
	assert.NoError(t, err)

	_, err = NewRedisStorage(
		context.Background(),
		cfg.Storages["myRedis"].Redis,
		domain.RealSystemClock,
		storagetesting.StubChannelProvider,
		telemetry.NewEmptyTelemetry(t),
	)
	assert.Regexp(t, `dial tcp 127.0.0.1:9999: connect: connection refused`, err.Error())
}

func TestInitialLoadScriptFailure(t *testing.T) {
	oldScript := publishMessageScript
	defer func() { publishMessageScript = oldScript }()
	publishMessageScript = redis.NewScript(`****`)

	cfg, err := config.ParseConfig(context.Background(), config.Overrides{}, fmt.Sprintf(`storages: { myRedis: { redis: { singleNode: "%s", connection: { max: 10 } } } }`, GetRedisAddr(t)))
	assert.NoError(t, err)

	_, err = NewRedisStorage(
		context.Background(),
		cfg.Storages["myRedis"].Redis,
		domain.RealSystemClock,
		storagetesting.StubChannelProvider,
		telemetry.NewEmptyTelemetry(t),
	)
	assert.Regexp(t, `Error compiling script \(new function\)`, err.Error())
}
