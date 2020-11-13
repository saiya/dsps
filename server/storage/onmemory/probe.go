package onmemory

import (
	"context"
)

func (s *onmemoryStorage) Liveness(ctx context.Context) (interface{}, error) {
	// If deadlock occurs, this check fails.
	unlock, err := s.lock.Lock(ctx)
	if err != nil {
		return nil, err
	}
	defer unlock()

	return "ok", nil
}

func (s *onmemoryStorage) Readiness(ctx context.Context) (interface{}, error) {
	// Onmemory storage is anytime ready
	return "ok", nil
}
