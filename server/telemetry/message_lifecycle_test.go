package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	ottrace "go.opentelemetry.io/otel/trace"
)

func TestInvalidMessageLifecycle(t *testing.T) {
	invalid := MessageLifecycle(-1)
	assert.Equal(t, "", invalid.otOperationName())
	assert.Equal(t, ottrace.SpanKindInternal, invalid.otSpanKind())
}
