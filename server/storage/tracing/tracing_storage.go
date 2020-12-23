package tracing

import (
	"context"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/telemetry"
)

type tracingStorage struct {
	id domain.StorageID
	t  *telemetry.Telemetry

	s      domain.Storage
	pubsub domain.PubSubStorage
	jwt    domain.JwtStorage
}

// NewTracingStorage wraps given Storage to trace calls
func NewTracingStorage(s domain.Storage, id domain.StorageID, telemetry *telemetry.Telemetry) domain.Storage {
	return &tracingStorage{
		id: id,
		t:  telemetry,

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
	ctx, end := ts.t.StartStorageSpan(ctx, ts.id, "Shutdown")
	defer end()
	return ts.s.Shutdown(ctx)
}

func (ts *tracingStorage) Liveness(ctx context.Context) (interface{}, error) {
	ctx, end := ts.t.StartStorageSpan(ctx, ts.id, "Liveness Probe")
	defer end()
	return ts.s.Liveness(ctx)
}

func (ts *tracingStorage) Readiness(ctx context.Context) (interface{}, error) {
	ctx, end := ts.t.StartStorageSpan(ctx, ts.id, "Readiness Probe")
	defer end()
	return ts.s.Readiness(ctx)
}
