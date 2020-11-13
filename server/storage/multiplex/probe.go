package multiplex

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/saiya/dsps/server/domain"
)

func (s *storageMultiplexer) Liveness(ctx context.Context) (interface{}, error) {
	return s.probe(ctx, "liveness", func(ctx context.Context, s domain.Storage) (interface{}, error) {
		return s.Liveness(ctx)
	})
}

func (s *storageMultiplexer) Readiness(ctx context.Context) (interface{}, error) {
	return s.probe(ctx, "readiness", func(ctx context.Context, s domain.Storage) (interface{}, error) {
		return s.Readiness(ctx)
	})
}

func (s *storageMultiplexer) probe(ctx context.Context, name string, delegate func(ctx context.Context, s domain.Storage) (interface{}, error)) (interface{}, error) {
	g, ctx := errgroup.WithContext(ctx)
	ch := make(chan childResult, len(s.children))
	for id, child := range s.children {
		id := id
		child := child
		g.Go(func() error {
			result, err := delegate(ctx, child)
			if err != nil {
				return fmt.Errorf("Storage \"%s\" %s check failed: %w", id, name, err)
			}
			ch <- childResult{
				id:    id,
				value: result,
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	close(ch)

	return struct {
		Children map[domain.StorageID]interface{} `json:"children"`
	}{
		Children: readAllChildResults(ch),
	}, nil
}
