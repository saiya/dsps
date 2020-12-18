package onmemory

import (
	"context"
	"errors"

	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/domain"
)

type onmemoryChannel struct {
	domain.Channel
	channelClock uint64

	subscribers map[domain.SubscriberID]*onmemorySubscriber
	log         map[domain.MessageLocator]*onmemoryMessage
}

type onmemorySubscriber struct {
	lastActivity domain.Time
	channelClock uint64
	messages     []*onmemoryMessage
}

func (s *onmemoryStorage) NewSubscriber(ctx context.Context, sl domain.SubscriberLocator) error {
	unlock, err := s.lock.Lock(ctx)
	if err != nil {
		return err
	}
	defer unlock()

	ch, err := s.getChannel(sl.ChannelID)
	if err != nil {
		return err
	}

	if ch.subscribers[sl.SubscriberID] != nil {
		return nil // Already exists (success)
	}

	instance := onmemorySubscriber{
		channelClock: ch.channelClock,
		lastActivity: s.systemClock.Now(),
		messages:     []*onmemoryMessage{},
	}
	ch.subscribers[sl.SubscriberID] = &instance
	return nil
}

func (s *onmemoryStorage) RemoveSubscriber(ctx context.Context, sl domain.SubscriberLocator) error {
	unlock, err := s.lock.Lock(ctx)
	if err != nil {
		return err
	}
	defer unlock()

	ch, err := s.getChannel(sl.ChannelID)
	if ch == nil && errors.Is(err, domain.ErrInvalidChannel) {
		// Because channel does not exist, subscriber also does not exist.
		// This method returns nil (success) if subscriber does not exist.
		return nil
	}
	if err != nil {
		return err
	}
	delete(ch.subscribers, sl.SubscriberID)
	return nil
}

func (s *onmemoryStorage) getChannel(id domain.ChannelID) (*onmemoryChannel, error) {
	ch := s.channels[id]
	if ch == nil {
		rawCh, err := s.channelProvider.Get(id)
		if err != nil {
			return nil, err
		}
		if rawCh != nil {
			ch = &onmemoryChannel{
				Channel:      rawCh,
				channelClock: 0,

				subscribers: map[domain.SubscriberID]*onmemorySubscriber{},
				log:         map[domain.MessageLocator]*onmemoryMessage{},
			}
			s.channels[id] = ch
		}
	}
	return ch, nil
}

// Note: this method holds lock of the storage!!
func (s *onmemoryStorage) findSubscriberForFetchMessages(ctx context.Context, sl domain.SubscriberLocator) (*onmemorySubscriber, error) {
	// This method is called from FetchMessages function.
	// It does not want to lock storage during polling.
	// So that lock storage in this method instead of the FetchMessages.
	unlock, err := s.lock.Lock(ctx)
	if err != nil {
		return nil, err
	}
	defer unlock()

	ch, err := s.getChannel(sl.ChannelID)
	if err != nil {
		return nil, err
	}

	sbsc := ch.subscribers[sl.SubscriberID]
	if sbsc == nil {
		return nil, xerrors.Errorf("%w", domain.ErrSubscriptionNotFound)
	}
	return sbsc, nil
}

func (sbsc *onmemorySubscriber) addMessage(msg onmemoryMessage) {
	sbsc.messages = append(sbsc.messages, &msg)
}
