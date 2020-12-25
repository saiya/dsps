package tracing_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/domain"
	. "github.com/saiya/dsps/server/domain/mock"
	. "github.com/saiya/dsps/server/storage/deps/testing"
	. "github.com/saiya/dsps/server/storage/tracing"
	"github.com/saiya/dsps/server/telemetry"
)

func withMockedStorage(t *testing.T, ctrl *gomock.Controller, f func(st domain.Storage, s *MockStorage)) {
	telemetry.WithStubTracing(t, func(telemetry *telemetry.Telemetry) {
		s := NewMockStorage(ctrl)
		s.EXPECT().AsPubSubStorage().Return(nil).Times(1)
		s.EXPECT().AsJwtStorage().Return(nil).Times(1)

		deps := EmptyDeps(t)
		deps.Telemetry = telemetry
		st := NewTracingStorage(s, "test", deps)
		f(st, s)
	})
}

func TestAsMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	withMockedStorage(t, ctrl, func(st domain.Storage, s *MockStorage) {
		assert.Nil(t, st.AsJwtStorage())
		assert.Nil(t, st.AsJwtStorage()) // Should cache inner storage result
		assert.Nil(t, st.AsPubSubStorage())
		assert.Nil(t, st.AsPubSubStorage())
	})
}

func TestPropertyPassthrough(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	withMockedStorage(t, ctrl, func(st domain.Storage, s *MockStorage) {
		s.EXPECT().String().Return("test stringer")
		assert.Equal(t, "test stringer", st.String())

		s.EXPECT().GetFileDescriptorPressure().Return(1234)
		assert.Equal(t, 1234, st.GetFileDescriptorPressure())
	})
}

func TestShutdownPassthrough(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	withMockedStorage(t, ctrl, func(st domain.Storage, s *MockStorage) {
		shutdownErr := errors.New("test error")
		s.EXPECT().Shutdown(gomock.Any()).Return(shutdownErr)
		assert.Same(t, shutdownErr, st.Shutdown(ctx))
	})
}

func TestProbePassthrough(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	withMockedStorage(t, ctrl, func(st domain.Storage, s *MockStorage) {
		var livenessResult interface{} = &struct{}{}
		livenessErr := errors.New("test error")
		s.EXPECT().Liveness(gomock.Any()).Return(livenessResult, livenessErr)
		result, err := st.Liveness(ctx)
		assert.Same(t, livenessResult, result)
		assert.Same(t, livenessErr, err)

		var readinessResult interface{} = &struct{}{}
		readinessErr := errors.New("test error")
		s.EXPECT().Readiness(gomock.Any()).Return(readinessResult, readinessErr)
		result, err = st.Readiness(ctx)
		assert.Same(t, readinessResult, result)
		assert.Same(t, readinessErr, err)
	})
}
