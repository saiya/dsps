package telemetry

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/label"
	ottrace "go.opentelemetry.io/otel/trace"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/logger"
)

// StartHTTPSpan starts tracing span for incoming or outgoing HTTP call
func (t *Telemetry) StartHTTPSpan(ctx context.Context, isServer bool, r *http.Request) (context.Context, context.CancelFunc) {
	var spanKind ottrace.SpanKind
	var requestURIlabel string
	if isServer {
		spanKind = ottrace.SpanKindServer
		requestURIlabel = "http.target" // For server request, RequestURI is domain relative URI
	} else {
		spanKind = ottrace.SpanKindClient // e.g. Outgoing webhook from DSPS to anywhere
		requestURIlabel = "http.url"
	}
	return t.startSpan(
		// > Therefore, HTTP client spans SHOULD be using conservative, low cardinality names formed from the available parameters of an HTTP request, such as "HTTP {METHOD_NAME}".
		// https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/trace/semantic_conventions/http.md
		ctx, fmt.Sprintf("HTTP %s", r.Method),
		ottrace.WithSpanKind(spanKind),
		ottrace.WithAttributes(
			// see: https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/trace/semantic_conventions/http.md
			label.String("http.method", r.Method),
			label.String("http.scheme", "http"), // FIXME:
			label.String("http.host", r.Host),
			label.String(requestURIlabel, r.RequestURI),
			label.String("http.user_agent", r.UserAgent()),
			label.Int64("http.request_content_length", r.ContentLength),
		),
	)
}

// SetHTTPServerAttributes adds attributes for incoming HTTP request
func (t *Telemetry) SetHTTPServerAttributes(ctx context.Context, r *http.Request, routePath string, clientIP string) {
	span := ottrace.SpanFromContext(ctx)
	span.SetName(fmt.Sprintf("HTTP %s %s", r.Method, routePath))
	span.SetAttributes(
		// https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/trace/semantic_conventions/http.md#http-server-semantic-conventions
		label.String("http.route", routePath),
		label.String("http.client_ip", clientIP),
	)
}

// SetHTTPResponseAttributes adds attributes of HTTP response
func (t *Telemetry) SetHTTPResponseAttributes(ctx context.Context, status int, contentLength int64) {
	ottrace.SpanFromContext(ctx).SetAttributes(
		label.Int("http.status_code", status),
		label.Int64("http.response_content_length", contentLength),
	)
}

// StartMessageSpan starts DSPS message processing span
func (t *Telemetry) StartMessageSpan(ctx context.Context, lifecycle MessageLifecycle, msg domain.Message) (context.Context, context.CancelFunc) {
	// see: https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/trace/semantic_conventions/messaging.md
	labels := []label.KeyValue{
		label.String("messaging.system", "dsps"),
		label.String("messaging.destination", string(msg.ChannelID)),
		label.String("messaging.message_id", string(msg.MessageID)),
		label.Int("messaging.message_payload_size_bytes", len([]byte(msg.Content))),
	}
	if lifecycle.otOperationName() != "send" {
		// > If the operation is "send", this attribute MUST NOT be set
		// https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/trace/semantic_conventions/messaging.md
		labels = append(labels, label.String("messaging.operation", lifecycle.otOperationName()))
	}
	return t.startSpan(
		// In DSPS, channel ID is high cardinality value and it is not easy to find single template name of it.
		// So that use simple span name as like as HTTP span name.
		ctx, fmt.Sprintf(`DSPS %s`, lifecycle.otOperationName()),
		ottrace.WithSpanKind(lifecycle.otSpanKind()),
		ottrace.WithAttributes(labels...),
	)
}

// StartStorageSpan starts storage operation span
func (t *Telemetry) StartStorageSpan(ctx context.Context, id domain.StorageID, operation string) (context.Context, context.CancelFunc) {
	return t.startSpan(
		ctx, fmt.Sprintf(`DSPS storage %s`, operation),
		ottrace.WithSpanKind(ottrace.SpanKindInternal),
		ottrace.WithAttributes(label.String("dsps.storage.id", string(id))),
	)
}

// SetSubscriberAttributes adds attributes of subscriber
func (t *Telemetry) SetSubscriberAttributes(ctx context.Context, sl domain.SubscriberLocator) {
	ottrace.SpanFromContext(ctx).SetAttributes(
		label.String("messaging.system", "dsps"),
		label.String("messaging.destination", string(sl.ChannelID)),
		label.String("dsps.subscriber_id", string(sl.SubscriberID)),
	)
}

// SetJTI adds attribute of JWT
func (t *Telemetry) SetJTI(ctx context.Context, jti domain.JwtJti) {
	ottrace.SpanFromContext(ctx).SetAttributes(
		label.String("jwt.jti", string(jti)),
	)
}

func (t *Telemetry) startSpan(ctx context.Context, name string, opts ...ottrace.SpanOption) (context.Context, context.CancelFunc) {
	var otSpan ottrace.Span
	ctx, otSpan = t.ot.Tracing.Tracer.Start(ctx, name, opts...)
	ctx = logger.PinLoggerContext(ctx) // without this, global error handler receives parent context that is not bound to span's context.
	return ctx, func() {
		if otSpan != nil {
			otSpan.End()
		}
	}
}
