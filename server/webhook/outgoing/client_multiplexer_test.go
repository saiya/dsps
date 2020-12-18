package outgoing

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/saiya/dsps/server/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewMultiplexClient(t *testing.T) {
	ctx := context.Background()

	handlerLock := sync.Mutex{}
	received := []domain.Message{}
	responseCodes := []int{}
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		bytes, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		var msg domain.Message
		assert.NoError(t, json.Unmarshal(bytes, &msg))

		handlerLock.Lock()
		defer handlerLock.Unlock()
		rw.WriteHeader(responseCodes[0])
		received = append(received, msg)
		responseCodes = responseCodes[1:]
	}))
	defer server.Close()

	tpl := newClientTemplateByConfig(t, `.+`, fmt.Sprintf(`{ "url": "%s", "retry": { "count": 1, "interval": "1ms", "intervalJitter": "1ms" } }`, server.URL))
	clients := make([]Client, 3)
	for i := range clients {
		var err error
		clients[i], err = tpl.NewClient(map[string]string{})
		assert.NoError(t, err)
	}
	multiplex := NewMultiplexClient(clients)
	defer multiplex.Close(ctx)

	assert.Regexp(t, server.URL, multiplex.String())

	// Webhook (success)
	msg := domain.Message{
		MessageLocator: domain.MessageLocator{ChannelID: "ch-1", MessageID: "webhook-1"},
		Content:        json.RawMessage(`{}`),
	}
	responseCodes = []int{204, 204, 204}
	received = []domain.Message{}
	assert.NoError(t, multiplex.Send(ctx, msg))
	assert.EqualValues(t, []domain.Message{msg, msg, msg}, received)

	// Webhook (retry + failure)
	responseCodes = []int{204, 500, 500, 500, 500} // Called 1+2+2 times due to retry
	received = []domain.Message{}
	assert.Regexp(t, `status code 500 returned`, multiplex.Send(ctx, msg).Error())
	assert.EqualValues(t, []domain.Message{msg, msg, msg, msg, msg}, received)
}

func TestEmptyMultiplex(t *testing.T) {
	ctx := context.Background()

	multiplex := NewMultiplexClient([]Client{})
	defer multiplex.Close(ctx)

	assert.NoError(t, multiplex.Send(ctx, domain.Message{
		MessageLocator: domain.MessageLocator{ChannelID: "ch-1", MessageID: "webhook-1"},
		Content:        json.RawMessage(`{}`),
	}))
}
