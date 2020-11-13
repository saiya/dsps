package redis

import (
	"context"
	"errors"

	"github.com/saiya/dsps/server/domain"
)

func (s *redisStorage) NewSubscriber(ctx context.Context, sl domain.SubscriberLocator) error {
	return errors.New("Not Implemented yet")
}

func (s *redisStorage) RemoveSubscriber(ctx context.Context, sl domain.SubscriberLocator) error {
	return errors.New("Not Implemented yet")
}
