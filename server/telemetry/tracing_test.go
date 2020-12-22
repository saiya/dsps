package telemetry

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	otsdktrace "go.opentelemetry.io/otel/sdk/trace"
	ottrace "go.opentelemetry.io/otel/trace"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/telemetry/opentelemetry"
	ottesting "github.com/saiya/dsps/server/telemetry/opentelemetry/testing"
)

func TestHTTPSpan(t *testing.T) {
	result := withStubTracing(t, func(t *Telemetry) {
		// Server request
		r := httptest.NewRequest("POST", "https://vhost.example.com/foo/bar", strings.NewReader("{}"))
		r.Header.Set("User-Agent", "test/0.1.2")
		ctx, close := t.StartHTTPSpan(context.Background(), true, r)
		t.SetHTTPServerAttributes(ctx, r, "/foo/{name}", "my.server.example.com", "172.0.0.2")
		t.SetHTTPResponseAttributes(ctx, 200, 123)
		close()

		// Client request
		r = httptest.NewRequest("GET", "http://vhost.example.com/outgoing/webhook", strings.NewReader("{}"))
		r.Header.Set("User-Agent", "test/0.1.2")
		ctx, close = t.StartHTTPSpan(context.Background(), false, r)
		t.SetHTTPResponseAttributes(ctx, 201, 123)
		close()
	})
	result.ot.AssertSpan(0, ottrace.SpanKindServer, "HTTP POST /foo/{name}", map[string]interface{}{
		// StartHTTPSpan
		"http.method":                 "POST",
		"http.url":                    "https://vhost.example.com/foo/bar",
		"http.host":                   "vhost.example.com",
		"http.scheme":                 "https",
		"http.user_agent":             "test/0.1.2",
		"http.request_content_length": int64(2),
		// SetHTTPServerAttributes
		"http.server_name": "my.server.example.com",
		"http.client_ip":   "172.0.0.2",
		// SetHTTPResponseAttributes
		"http.status_code":             int64(200),
		"http.response_content_length": int64(123),
	})
	result.ot.AssertSpan(1, ottrace.SpanKindClient, "HTTP GET", map[string]interface{}{
		// StartHTTPSpan
		"http.method":                 "GET",
		"http.url":                    "http://vhost.example.com/outgoing/webhook",
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
	result := withStubTracing(t, func(t *Telemetry) {
		_, close := t.StartMessageSpan(context.Background(), Send, msg)
		close()
		_, close = t.StartMessageSpan(context.Background(), Fetch, msg)
		close()
		_, close = t.StartMessageSpan(context.Background(), Acknowledge, msg)
		close()
	})
	result.ot.AssertSpan(0, ottrace.SpanKindProducer, "DSPS send", map[string]interface{}{
		"messaging.system":                     "dsps",
		"messaging.message_id":                 string(msg.MessageID),
		"messaging.conversation_id":            string(msg.ChannelID),
		"messaging.message_payload_size_bytes": int64(len(msg.Content)),
	})
	result.ot.AssertSpan(1, ottrace.SpanKindConsumer, "DSPS receive", map[string]interface{}{
		"messaging.system":                     "dsps",
		"messaging.message_id":                 string(msg.MessageID),
		"messaging.conversation_id":            string(msg.ChannelID),
		"messaging.message_payload_size_bytes": int64(len(msg.Content)),
		"messaging.operation":                  "receive",
	})
	result.ot.AssertSpan(2, ottrace.SpanKindConsumer, "DSPS process", map[string]interface{}{
		"messaging.system":                     "dsps",
		"messaging.message_id":                 string(msg.MessageID),
		"messaging.conversation_id":            string(msg.ChannelID),
		"messaging.message_payload_size_bytes": int64(len(msg.Content)),
		"messaging.operation":                  "process",
	})
}

func TestStorageSpan(t *testing.T) {
	result := withStubTracing(t, func(t *Telemetry) {
		_, close := t.StartStorageSpan(context.Background(), "storage-1", "DoSomething")
		close()
	})
	result.ot.AssertSpan(0, ottrace.SpanKindInternal, "DSPS storage DoSomething", map[string]interface{}{
		"dsps.storage.id": "storage-1",
	})
}

type traceResult struct {
	ot *ottesting.StubExporter
}

func withStubTracing(t *testing.T, f func(*Telemetry)) *traceResult {
	tr := &traceResult{
		ot: ottesting.NewStubExporter(t),
	}
	telemetry := &Telemetry{
		ot: &opentelemetry.OTFacility{
			Tracing: &opentelemetry.Tracing{
				Tracer: otsdktrace.NewTracerProvider(otsdktrace.WithSyncer(tr.ot)).Tracer("test"),
			},
		},
	}
	defer telemetry.Shutdown(context.Background())
	f(telemetry)
	return tr
}
