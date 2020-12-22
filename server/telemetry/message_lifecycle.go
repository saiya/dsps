package telemetry

import (
	ottrace "go.opentelemetry.io/otel/trace"
)

// MessageLifecycle of the domain.Message (in tracing aspect).
type MessageLifecycle int

const (
	// Send means sending message
	Send MessageLifecycle = iota
	// Fetch means client fetching
	Fetch
	// Acknowledge means client received span
	Acknowledge
)

// https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/trace/semantic_conventions/messaging.md
func (lifecycle MessageLifecycle) otOperationName() string {
	switch lifecycle {
	case Send:
		return "send"
	case Fetch:
		return "receive"
	case Acknowledge:
		return "process"
	default:
		return ""
	}
}

func (lifecycle MessageLifecycle) otSpanKind() ottrace.SpanKind {
	switch lifecycle {
	case Send:
		return ottrace.SpanKindProducer
	case Fetch:
		return ottrace.SpanKindConsumer
	case Acknowledge:
		return ottrace.SpanKindConsumer
	default:
		return ottrace.SpanKindInternal
	}
}
