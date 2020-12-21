package tracing

import (
	"context"

	"github.com/saiya/dsps/server/domain"
)

type tracingStorage struct {
	s      domain.Storage
	pubsub domain.PubSubStorage
	jwt    domain.JwtStorage
}

// NewTracingStorage wraps given Storage to trace calls
func NewTracingStorage(s domain.Storage, id domain.StorageID) domain.Storage {
	return &tracingStorage{
		s:      s,
		pubsub: s.AsPubSubStorage(),
		jwt:    s.AsJwtStorage(),
	}
}

func (ts *tracingStorage) AsPubSubStorage() domain.PubSubStorage {
	if ts.pubsub == nil {
		return nil
	}
	return ts
}

func (ts *tracingStorage) AsJwtStorage() domain.JwtStorage {
	if ts.jwt == nil {
		return nil
	}
	return ts
}

func (ts *tracingStorage) String() string {
	return ts.s.String()
}

func (ts *tracingStorage) GetFileDescriptorPressure() int {
	return ts.s.GetFileDescriptorPressure()
}

func (ts *tracingStorage) Shutdown(ctx context.Context) error {
	return ts.s.Shutdown(ctx)
}

func (ts *tracingStorage) Liveness(ctx context.Context) (interface{}, error) {
	return ts.s.Liveness(ctx)
}

func (ts *tracingStorage) Readiness(ctx context.Context) (interface{}, error) {
	return ts.s.Readiness(ctx)
}
