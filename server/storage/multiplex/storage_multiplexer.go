package multiplex

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/saiya/dsps/server/domain"
)

// NewStorageMultiplexer creates Storage instance that wraps multiple Storage instances
func NewStorageMultiplexer(children map[domain.StorageID]domain.Storage) (domain.Storage, error) {
	if len(children) == 0 {
		return nil, fmt.Errorf("List of storages must not be empty")
	}

	pubsubSupported := false
	jwtSupported := false
	for _, c := range children {
		if pubsub := c.AsPubSubStorage(); pubsub != nil {
			pubsubSupported = true
		}
		if jwt := c.AsJwtStorage(); jwt != nil {
			jwtSupported = true
		}
	}

	return &storageMultiplexer{
		children: children,

		pubsubSupported: pubsubSupported,
		jwtSupported:    jwtSupported,
	}, nil
}

type storageMultiplexer struct {
	children map[domain.StorageID]domain.Storage

	pubsubSupported bool
	jwtSupported    bool
}

func (s *storageMultiplexer) AsPubSubStorage() domain.PubSubStorage {
	if !s.pubsubSupported {
		return nil
	}
	return s
}
func (s *storageMultiplexer) AsJwtStorage() domain.JwtStorage {
	if !s.jwtSupported {
		return nil
	}
	return s
}

func (s *storageMultiplexer) String() string {
	return storageMapToString(s.children)
}

func storageMapToString(children map[domain.StorageID]domain.Storage) string {
	list := make([]string, 0, len(children))
	for _, c := range children {
		list = append(list, c.String())
	}
	return strings.Join(list, ",")
}

func (s *storageMultiplexer) Shutdown(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	for _, c := range s.children {
		child := c
		g.Go(func() error {
			return child.Shutdown(ctx)
		})
	}
	return g.Wait()
}

func (s *storageMultiplexer) GetNoFilePressure() int {
	result := 0
	for _, c := range s.children {
		result += c.GetNoFilePressure()
	}
	return result
}
