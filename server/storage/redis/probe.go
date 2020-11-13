package redis

import (
	"context"
	"errors"
)

func (s *redisStorage) Liveness(ctx context.Context) (interface{}, error) {
	return map[string]string{}, errors.New("Not Implemented yet")
}

func (s *redisStorage) Readiness(ctx context.Context) (interface{}, error) {
	return map[string]string{}, errors.New("Not Implemented yet")
}
