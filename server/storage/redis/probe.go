package redis

import (
	"context"
)

func (s *redisStorage) Liveness(ctx context.Context) (interface{}, error) {
	return map[string]string{}, nil
}

func (s *redisStorage) Readiness(ctx context.Context) (interface{}, error) {
	if err := s.redisCmd.Ping(ctx); err != nil {
		return nil, err
	}
	return map[string]string{"redis": "ping OK"}, nil
}
