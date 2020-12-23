package testing

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	otlabel "go.opentelemetry.io/otel/label"
	ottraceexport "go.opentelemetry.io/otel/sdk/export/trace"
	ottrace "go.opentelemetry.io/otel/trace"
)

// StubExporter implements OT SpanExporter
type StubExporter struct {
	t *testing.T

	lock    sync.Mutex
	otSpans []*ottraceexport.SpanData
}

// NewStubExporter creates new instance
func NewStubExporter(t *testing.T) *StubExporter {
	return &StubExporter{t: t}
}

// AssertSpan asserts specified span captured
func (tr *StubExporter) AssertSpan(index int, kind ottrace.SpanKind, name string, attributes map[string]interface{}) *ottraceexport.SpanData {
	tr.lock.Lock()
	defer tr.lock.Unlock()

	t := tr.t
	if !assert.Greater(t, len(tr.otSpans), index) {
		return nil
	}

	span := tr.otSpans[index]
	assert.Equal(t, kind, span.SpanKind)
	assert.Equal(t, name, span.Name)
	assert.EqualValues(t, attributes, KVPairsToMap(span.Attributes))
	return span
}

// GetSpans returns captured spans
func (tr *StubExporter) GetSpans() []*ottraceexport.SpanData {
	tr.lock.Lock()
	defer tr.lock.Unlock()

	return tr.otSpans
}

// ExportSpans implements OT SpanExporter
func (tr *StubExporter) ExportSpans(ctx context.Context, spanData []*ottraceexport.SpanData) error {
	tr.lock.Lock()
	defer tr.lock.Unlock()

	tr.otSpans = append(tr.otSpans, spanData...)
	return nil
}

// Shutdown implements OT SpanExporter
func (tr *StubExporter) Shutdown(ctx context.Context) error {
	return nil
}

// KVPairsToMap convert OT labels to map
func KVPairsToMap(pairs []otlabel.KeyValue) map[string]interface{} {
	result := make(map[string]interface{}, len(pairs))
	for _, pair := range pairs {
		result[string(pair.Key)] = pair.Value.AsInterface()
	}
	return result
}
