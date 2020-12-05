package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/domain"
)

func randomChannelID(_ *testing.T) domain.ChannelID {
	uuid, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	return domain.ChannelID(fmt.Sprintf("ch-%s", uuid))
}

func assertValueAndTTL(t *testing.T, redisCmd redisCmd, key string, value string, ttl time.Duration) bool {
	actual, err := redisCmd.Get(context.Background(), key)
	assert.NoError(t, err)
	return assert.Equal(t, &value, actual) && assertTTL(t, redisCmd, key, ttl)
}

func assertTTL(t *testing.T, redisCmd redisCmd, key string, ttl time.Duration) bool {
	ctx := context.Background()
	actual, err := redisCmd.TTL(ctx, key)
	assert.NoError(t, err)
	return assert.LessOrEqual(t, actual.Seconds(), ttl.Seconds()) &&
		assert.LessOrEqual(t, ttl.Seconds()-1, actual.Seconds())
}

func strPList(_ *testing.T, values ...string) []*string {
	result := make([]*string, len(values))
	for i, value := range values {
		tmp := value
		result[i] = &tmp
	}
	return result
}
