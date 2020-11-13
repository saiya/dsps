package multiplex

import (
	"context"

	"github.com/saiya/dsps/server/domain"
)

func (s *storageMultiplexer) NewSubscriber(ctx context.Context, sl domain.SubscriberLocator) error {
	_, err := s.parallelAtLeastOneSuccess(ctx, "NewSubscriber", func(ctx context.Context, _ domain.StorageID, child domain.Storage) (interface{}, error) {
		if child := child.AsPubSubStorage(); child != nil {
			return nil, child.NewSubscriber(ctx, sl)
		}
		return nil, errMultiplexSkipped
	})
	return err
}

func (s *storageMultiplexer) RemoveSubscriber(ctx context.Context, sl domain.SubscriberLocator) error {
	_, err := s.parallelAtLeastOneSuccess(ctx, "RemoveSubscriber", func(ctx context.Context, _ domain.StorageID, child domain.Storage) (interface{}, error) {
		if child := child.AsPubSubStorage(); child != nil {
			return nil, child.RemoveSubscriber(ctx, sl)
		}
		return nil, errMultiplexSkipped
	})
	return err
}
