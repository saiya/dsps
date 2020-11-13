package redis

import (
	"context"
)

type redisStorageStat struct {
}

func (s *redisStorage) Stat(ctx context.Context) (interface{}, error) {
	snapshot := *s.stat
	return &snapshot, nil
}
