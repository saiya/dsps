package telemetry

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	ottrace "go.opentelemetry.io/otel/trace"

	"github.com/saiya/dsps/server/domain"
)

func TestHTTPSpan(t *testing.T) {
	result := WithStubTracing(t, func(t *Telemetry) {
		// Server request
		r := httptest.NewRequest("POST", "/foo/bar?param=value", strings.NewReader("{}"))
		r.Host = "vhost.example.com"
		r.Header.Set("User-Agent", "test/0.1.2")
		ctx, close := t.StartHTTPSpan(context.Background(), true, r)
		t.SetHTTPServerAttributes(ctx, r, "/foo/{name}", "172.0.0.2")
		t.SetHTTPResponseAttributes(ctx, 200, 123)
		close()

		// Client request
		r = httptest.NewRequest("GET", "http://vhost.example.com/outgoing/webhook?param=value", strings.NewReader("{}"))
		r.Header.Set("User-Agent", "test/0.1.2")
		ctx, close = t.StartHTTPSpan(context.Background(), false, r)
		t.SetHTTPResponseAttributes(ctx, 201, 123)
		close()
	})
	result.OT.AssertSpan(0, ottrace.SpanKindServer, "HTTP POST /foo/{name}", map[string]interface{}{
		// StartHTTPSpan
		"http.method":                 "POST",
		"http.scheme":                 "http",
		"http.host":                   "vhost.example.com",
		"http.target":                 "/foo/bar?param=value",
		"http.user_agent":             "test/0.1.2",
		"http.request_content_length": int64(2),
		// SetHTTPServerAttributes
		"http.route":     "/foo/{name}",
		"http.client_ip": "172.0.0.2",
		// SetHTTPResponseAttributes
		"http.status_code":             int64(200),
		"http.response_content_length": int64(123),
	})
	result.OT.AssertSpan(1, ottrace.SpanKindClient, "HTTP GET vhost.example.com", map[string]interface{}{
		// StartHTTPSpan
		"http.method":                 "GET",
		"http.url":                    "http://vhost.example.com/outgoing/webhook?param=value",
		"http.host":                   "vhost.example.com",
		"http.scheme":                 "http",
		"http.user_agent":             "test/0.1.2",
		"http.request_content_length": int64(2),
		// SetHTTPResponseAttributes
		"http.status_code":             int64(201),
		"http.response_content_length": int64(123),
	})
}

func TestMessageSpan(t *testing.T) {
	msg := domain.Message{
		MessageLocator: domain.MessageLocator{
			ChannelID: "ch-1",
			MessageID: "msg-1",
		},
		Content: json.RawMessage("{}"),
	}
	result := WithStubTracing(t, func(t *Telemetry) {
		_, close := t.StartMessageSpan(context.Background(), Send, msg)
		close()
		_, close = t.StartMessageSpan(context.Background(), Fetch, msg)
		close()
		_, close = t.StartMessageSpan(context.Background(), Acknowledge, msg)
		close()
	})
	result.OT.AssertSpan(0, ottrace.SpanKindProducer, "DSPS send", map[string]interface{}{
		"messaging.system":                     "dsps",
		"messaging.message_id":                 string(msg.MessageID),
		"messaging.destination":                string(msg.ChannelID),
		"messaging.message_payload_size_bytes": int64(len(msg.Content)),
	})
	result.OT.AssertSpan(1, ottrace.SpanKindConsumer, "DSPS receive", map[string]interface{}{
		"messaging.system":                     "dsps",
		"messaging.message_id":                 string(msg.MessageID),
		"messaging.destination":                string(msg.ChannelID),
		"messaging.message_payload_size_bytes": int64(len(msg.Content)),
		"messaging.operation":                  "receive",
	})
	result.OT.AssertSpan(2, ottrace.SpanKindConsumer, "DSPS process", map[string]interface{}{
		"messaging.system":                     "dsps",
		"messaging.message_id":                 string(msg.MessageID),
		"messaging.destination":                string(msg.ChannelID),
		"messaging.message_payload_size_bytes": int64(len(msg.Content)),
		"messaging.operation":                  "process",
	})
}

func TestStorageSpanWithMessagingAttrs(t *testing.T) {
	result := WithStubTracing(t, func(t *Telemetry) {
		ctx, close := t.StartStorageSpan(context.Background(), "storage-1", "DoSomething")
		t.SetSubscriberAttributes(ctx, domain.SubscriberLocator{
			ChannelID:    "ch-1",
			SubscriberID: "sbsc-1",
		})
		close()
	})
	result.OT.AssertSpan(0, ottrace.SpanKindInternal, "DSPS storage DoSomething", map[string]interface{}{
		"dsps.storage.id":       "storage-1",
		"messaging.system":      "dsps",
		"messaging.destination": "ch-1",
		"dsps.subscriber_id":    "sbsc-1",
	})
}

func TestJTIAttrs(t *testing.T) {
	result := WithStubTracing(t, func(t *Telemetry) {
		ctx, close := t.StartStorageSpan(context.Background(), "storage-1", "DoSomething")
		t.SetJTI(ctx, "jti-value")
		close()

	})
	result.OT.AssertSpan(0, ottrace.SpanKindInternal, "DSPS storage DoSomething", map[string]interface{}{
		"dsps.storage.id": "storage-1",
		"jwt.jti":         "jti-value",
	})
}

func TestDaemonSpan(t *testing.T) {
	result := WithStubTracing(t, func(t *Telemetry) {
		_, close := t.StartDaemonSpan(context.Background(), "test.daemon", "doSomething")
		close()
	})
	result.OT.AssertSpan(0, ottrace.SpanKindInternal, "BackgroundJob test.daemon doSomething", map[string]interface{}{
		"dsps.daemon.system": "test.daemon",
		"dsps.daemon.name":   "doSomething",
	})
}
