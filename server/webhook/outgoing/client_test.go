package outgoing

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/sentry"
	"github.com/saiya/dsps/server/telemetry"
)

func newClientAndServerByConfig(t *testing.T, handler http.Handler, tplEnv domain.TemplateStringEnv, config string, h func(client *clientImpl)) {
	server := httptest.NewServer(handler)
	defer server.Close()

	tpl := newClientTemplateByConfig(t, `.+`, strings.ReplaceAll(config, "${BASE_URL}", server.URL))
	tpl.Close()
	assert.NotNil(t, tpl)

	client, err := tpl.NewClient(tplEnv)
	assert.NoError(t, err)
	defer client.Close(context.Background())

	if h != nil {
		h(client.(*clientImpl))
	}
}

func TestClientImpl(t *testing.T) {
	msg := domain.Message{
		MessageLocator: domain.MessageLocator{
			ChannelID: "chat-room-1234",
			MessageID: "msg-1",
		},
		Content: []byte(`{"hi":"hello"}`),
	}
	var received domain.Message
	var handler http.HandlerFunc = func(rw http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, `/you-got-message/room/1234`, r.URL.Path)
		assert.Equal(t, `My DSPS server`, r.Header.Get("User-Agent"))
		assert.Equal(t, `1234`, r.Header.Get("X-Room-ID"))

		body, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%d", len(body)), r.Header.Get("Content-Length"))
		assert.NoError(t, json.Unmarshal(body, &received))
	}
	newClientAndServerByConfig(
		t,
		handler,
		map[string]interface{}{"channel": map[string]string{"id": "1234"}},
		`{
			"method": "PUT",
			"url": "${BASE_URL}/you-got-message/room/{{.channel.id}}", 
			"timeout": "3s",
			"headers": { 
				"User-Agent": "My DSPS server",
				"X-Room-ID": "{{.channel.id}}"
			}
		}`,
		func(client *clientImpl) {
			assert.NoError(t, client.Send(context.Background(), msg))
		},
	)
	assert.EqualValues(t, msg, received)
}

func TestClientTracing(t *testing.T) {
	msg := domain.Message{
		MessageLocator: domain.MessageLocator{
			ChannelID: "chat-room-1234",
			MessageID: "msg-1",
		},
		Content: []byte(`{"hi":"hello"}`),
	}
	var received domain.Message
	var handler http.HandlerFunc = func(rw http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, `/you-got-message/room/1234`, r.URL.Path)
		assert.Equal(t, `My DSPS server`, r.Header.Get("User-Agent"))
		assert.Equal(t, `1234`, r.Header.Get("X-Room-ID"))

		body, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%d", len(body)), r.Header.Get("Content-Length"))
		assert.NoError(t, json.Unmarshal(body, &received))
	}
	var serverHostName string
	tr := telemetry.WithStubTracing(t, func(telemetry *telemetry.Telemetry) {
		newClientAndServerByConfig(
			t,
			handler,
			map[string]interface{}{"channel": map[string]string{"id": "1234"}},
			`{
				"method": "PUT",
				"url": "${BASE_URL}/you-got-message/room/{{.channel.id}}", 
				"timeout": "3s",
				"headers": { 
					"User-Agent": "My DSPS server",
					"X-Room-ID": "{{.channel.id}}"
				}
			}`,
			func(client *clientImpl) {
				url, err := url.Parse(client.url)
				assert.NoError(t, err)
				serverHostName = url.Host

				client.telemetry = telemetry
				assert.NoError(t, client.Send(context.Background(), msg))
			},
		)
	})
	assert.EqualValues(t, msg, received)
	tr.OT.AssertSpan(0, trace.SpanKindClient, "HTTP PUT "+serverHostName, map[string]interface{}{
		"http.method":                  "PUT",
		"http.scheme":                  "http",
		"http.host":                    serverHostName,
		"http.url":                     fmt.Sprintf("http://%s/you-got-message/room/1234", serverHostName),
		"http.user_agent":              "My DSPS server",
		"http.request_content_length":  int64(114),
		"http.status_code":             int64(200),
		"http.response_content_length": int64(0),
	})
}

func TestClientRetry(t *testing.T) {
	msg := domain.Message{
		MessageLocator: domain.MessageLocator{
			ChannelID: "chat-room-1234",
			MessageID: "msg-1",
		},
		Content: []byte(`{"hi":"hello"}`),
	}
	var received domain.Message
	handlerCalled := 0
	var handler http.HandlerFunc = func(rw http.ResponseWriter, r *http.Request) {
		handlerCalled++
		if handlerCalled <= 3 {
			rw.WriteHeader(500)
			return
		}

		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, `/you-got-message/room/1234`, r.URL.Path)

		body, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.NoError(t, json.Unmarshal(body, &received))
	}
	sentry := sentry.NewStubSentry()
	newClientAndServerByConfig(
		t,
		handler,
		map[string]interface{}{"channel": map[string]string{"id": "1234"}},
		`{
			"url": "${BASE_URL}/you-got-message/room/{{.channel.id}}", 
			"retry": {
				"count": 3,
				"interval": "1ms",
				"intervalJitter": "1ms"
			}
		}`,
		func(client *clientImpl) {
			ctx := sentry.WrapContext(context.Background())
			assert.NoError(t, client.Send(ctx, msg))
		},
	)
	assert.EqualValues(t, msg, received)
	assert.Equal(t, 1+3, handlerCalled)

	breadcrumbs := sentry.GetBreadcrumbs()
	assert.Equal(t, 4, len(breadcrumbs))
	assert.Equal(t, "http", breadcrumbs[0].Type)
	assert.Equal(t, "Outgoing webhook", breadcrumbs[0].Message)
	assert.Equal(t, "PUT", breadcrumbs[0].Data["method"])
	assert.Equal(t, 500, breadcrumbs[0].Data["status_code"])
	assert.Regexp(t, "http://127.0.0.1:[0-9]+/you-got-message/room/1234", breadcrumbs[0].Data["url"].(fmt.Stringer).String())
}

func TestClientRetryFailure(t *testing.T) {
	msg := domain.Message{
		MessageLocator: domain.MessageLocator{
			ChannelID: "chat-room-1234",
			MessageID: "msg-1",
		},
		Content: []byte(`{"hi":"hello"}`),
	}
	handlerCalled := 0
	var handler http.HandlerFunc = func(rw http.ResponseWriter, r *http.Request) {
		handlerCalled++
		rw.WriteHeader(500)
	}
	sentry := sentry.NewStubSentry()
	newClientAndServerByConfig(
		t,
		handler,
		map[string]interface{}{"channel": map[string]string{"id": "1234"}},
		`{
			"url": "${BASE_URL}/you-got-message/room/{{.channel.id}}", 
			"retry": {
				"count": 3,
				"interval": "1ms",
				"intervalJitter": "1ms"
			}
		}`,
		func(client *clientImpl) {
			ctx := sentry.WrapContext(context.Background())
			assert.Regexp(t, `status code 500 returned`, client.Send(ctx, msg).Error())
		},
	)
	assert.Equal(t, 1+3, handlerCalled)
	assert.Regexp(t, `outgoing webhook failed: status code 500 returned`, sentry.GetLastError())
}

func TestClientInvalidUrl(t *testing.T) {
	tpl := newClientTemplateByConfig(t, `.+`, `{ "url": "://example.com", "retry": { "interval": "1ms", "intervalJitter": "1ms" } }`)
	tpl.Close()
	assert.NotNil(t, tpl)

	client, err := tpl.NewClient(map[string]interface{}{})
	assert.NoError(t, err)

	assert.Regexp(t, `missing protocol scheme`, client.Send(context.Background(), domain.Message{}).Error())
}

func TestClientClose(t *testing.T) {
	tpl := newClientTemplateByConfig(t, `.+`, strings.ReplaceAll(`{ "url": "${BASE_URL}/you-got-message/room/1234" }`, "${BASE_URL}", "http://localhost:1234"))
	tpl.Close()
	assert.NotNil(t, tpl)

	client, err := tpl.NewClient(map[string]interface{}{})
	assert.NoError(t, err)

	client.Close(context.Background())
	client.Close(context.Background())
	assert.Regexp(t, `outgoing-webhook client already closed`, client.Send(context.Background(), domain.Message{}).Error())
}

func TestCorruptedMessageContent(t *testing.T) {
	newClientAndServerByConfig(
		t,
		nil,
		nil,
		`{ "url": "${BASE_URL}/you-got-message/room/1234" }`,
		func(client *clientImpl) {
			assert.Regexp(t, `failed to make request body of outgoing-webhook`, client.Send(context.Background(), domain.Message{
				MessageLocator: domain.MessageLocator{
					ChannelID: "ch-1",
					MessageID: "msg-1",
				},
				Content: json.RawMessage(`{{{`),
			}).Error())
		},
	)
}

func TestClientTemplateEvalFailures(t *testing.T) {
	tpl := newClientTemplateByConfig(t, `.+`, strings.ReplaceAll(`{
		"url": "${BASE_URL}/you-got-message/room/{{.INVALID}}",
	}`, "${BASE_URL}", "http://localhost:1234"))
	tpl.Close()
	assert.NotNil(t, tpl)
	_, err := tpl.NewClient(map[string]interface{}{})
	assert.Regexp(t, `map has no entry for key "INVALID"`, err.Error())

	tpl = newClientTemplateByConfig(t, `.+`, strings.ReplaceAll(`{
		"url": "${BASE_URL}/you-got-message/room/1234",
		"headers": { "X-Something": "{{.INVALID}}" }
	}`, "${BASE_URL}", "http://localhost:1234"))
	tpl.Close()
	assert.NotNil(t, tpl)
	_, err = tpl.NewClient(map[string]interface{}{})
	assert.Regexp(t, `map has no entry for key "INVALID"`, err.Error())
}
