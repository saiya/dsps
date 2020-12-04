package onmemory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	. "github.com/saiya/dsps/server/storage/testing"
)

func makeRawStorage(t *testing.T) *onmemoryStorage {
	s, err := NewOnmemoryStorage(context.Background(), &config.OnmemoryStorageConfig{}, domain.RealSystemClock, StubChannelProvider)
	if !assert.NoError(t, err) {
		return nil
	}

	raw, ok := s.(*onmemoryStorage)
	if !ok {
		assert.FailNow(t, "cast failed")
	}
	return raw
}
