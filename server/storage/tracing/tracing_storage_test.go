package tracing_test

import (
	"context"
	"testing"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/storage/onmemory"
	. "github.com/saiya/dsps/server/storage/testing"
	. "github.com/saiya/dsps/server/storage/tracing"
	"github.com/saiya/dsps/server/telemetry"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

var onmemoryTracingCtor = func(telemetry *telemetry.Telemetry, onmemConfig config.OnmemoryStorageConfig) StorageCtor {
	return func(ctx context.Context, systemClock domain.SystemClock, channelProvider domain.ChannelProvider) (domain.Storage, error) {
		storage, err := onmemory.NewOnmemoryStorage(context.Background(), &onmemConfig, systemClock, channelProvider)
		if err != nil {
			return nil, err
		}
		return NewTracingStorage(storage, "test", telemetry), nil
	}
}

func testTracing(t *testing.T, f func(domain.Storage)) *telemetry.TraceResult {
	return telemetry.WithStubTracing(t, func(telemetry *telemetry.Telemetry) {
		s, err := onmemoryTracingCtor(telemetry, config.OnmemoryStorageConfig{})(
			context.Background(),
			domain.RealSystemClock,
			StubChannelProvider,
		)
		assert.NoError(t, err)

		f(s)
		assert.NoError(t, s.Shutdown(context.Background()))
	})
}

func TestCoreFunctionsTrace(t *testing.T) {
	tr := testTracing(t, func(s domain.Storage) {
		_, err := s.Liveness(context.Background())
		assert.NoError(t, err)
		_, err = s.Readiness(context.Background())
		assert.NoError(t, err)
	})
	tr.OT.AssertSpan(0, trace.SpanKindInternal, "DSPS storage Liveness Probe", map[string]interface{}{
		"dsps.storage.id": "test",
	})
	tr.OT.AssertSpan(1, trace.SpanKindInternal, "DSPS storage Readiness Probe", map[string]interface{}{
		"dsps.storage.id": "test",
	})
	tr.OT.AssertSpan(2, trace.SpanKindInternal, "DSPS storage Shutdown", map[string]interface{}{
		"dsps.storage.id": "test",
	})
}

func TestCoreFunction(t *testing.T) {
	telemetry.WithStubTracing(t, func(telemetry *telemetry.Telemetry) {
		CoreFunctionTest(t, onmemoryTracingCtor(telemetry, config.OnmemoryStorageConfig{
			DisableJwt:    true,
			DisablePubSub: true,
		}))
	})
}

func TestPubSub(t *testing.T) {
	telemetry.WithStubTracing(t, func(telemetry *telemetry.Telemetry) {
		PubSubTest(t, onmemoryTracingCtor(telemetry, config.OnmemoryStorageConfig{
			DisableJwt: true,
		}))
	})
}

func TestJwt(t *testing.T) {
	telemetry.WithStubTracing(t, func(telemetry *telemetry.Telemetry) {
		JwtTest(t, onmemoryTracingCtor(telemetry, config.OnmemoryStorageConfig{
			DisablePubSub: true,
		}))
	})
}
