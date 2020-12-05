package testing

import (
	"context"
	"fmt"
	"testing"
	"time"

	domain "github.com/saiya/dsps/server/domain"
)

// StorageCtor should construct Storage instance to test
type StorageCtor func(ctx context.Context, systemClock domain.SystemClock, channelProvider domain.ChannelProvider) (domain.Storage, error)

func storageSubTest(t *testing.T, storageCtor StorageCtor, name string, f func(t *testing.T, storageCtor StorageCtor)) {
	startAt := time.Now()
	fmt.Printf("          RUN %s\n", name)
	f(t, storageCtor)
	fmt.Printf("          END %s (%s)\n", name, time.Since(startAt).String())
}
