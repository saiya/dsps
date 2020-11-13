package redis

import (
	"context"
	"errors"

	"github.com/saiya/dsps/server/domain"
)

func (s *redisStorage) PublishMessages(ctx context.Context, msgs []domain.Message) error {
	return errors.New("Not Implemented yet")
}

func (s *redisStorage) FetchMessages(ctx context.Context, sl domain.SubscriberLocator, max int, waituntil domain.Duration) (messages []domain.Message, moreMessages bool, ackHandle domain.AckHandle, err error) {
	return []domain.Message{}, false, domain.AckHandle{}, errors.New("Not Implemented yet")
}

func (s *redisStorage) AcknowledgeMessages(ctx context.Context, handle domain.AckHandle) error {
	return errors.New("Not Implemented yet")
}

func (s *redisStorage) IsOldMessages(ctx context.Context, sl domain.SubscriberLocator, msgs []domain.MessageLocator) (map[domain.MessageLocator]bool, error) {
	return nil, errors.New("Not Implemented yet")
}
