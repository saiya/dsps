package outgoing

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/logger"
)

type multiplexClient struct {
	clients []Client
}

// NewMultiplexClient wraps given clients as a single client
func NewMultiplexClient(clients []Client) Client {
	return &multiplexClient{clients: clients}
}

func (mux *multiplexClient) String() string {
	result := make([]string, len(mux.clients))
	for i, c := range mux.clients {
		result[i] = c.String()
	}
	return strings.Join(result, ", ")
}

func (mux *multiplexClient) Send(ctx context.Context, msg domain.Message) error {
	var lastError error = nil
	for _, c := range mux.clients {
		if err := c.Send(ctx, msg); err != nil {
			if lastError != nil {
				logger.Of(ctx).WarnError(logger.CatOutgoingWebhook, "multiple error on outgoing-webhook multiplexer", lastError)
			}
			lastError = fmt.Errorf("%s: %w", c.String(), err)
		}
	}
	return lastError
}

func (mux *multiplexClient) Close(ctx context.Context) {
	wg := sync.WaitGroup{}
	for i := range mux.clients {
		wg.Add(1)
		c := mux.clients[i]
		go func() {
			defer wg.Done()
			c.Close(ctx)
		}()
	}
	wg.Wait()
}
