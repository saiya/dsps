package onmemory

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"time"

	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/domain"
)

type onmemoryMessage struct {
	domain.Message
	channelClock uint64
	ExpireAt     domain.Time
}

func (msg *onmemoryMessage) Validate() error {
	if _, err := json.Marshal(msg.Message.Content); err != nil {
		return xerrors.Errorf("%w: %v", domain.ErrMalformedMessageJSON, err)
	}
	return nil
}

func (s *onmemoryStorage) PublishMessages(ctx context.Context, msgs []domain.Message) error {
	if !domain.BelongsToSameChannel(msgs) {
		return xerrors.New("Messages belongs to various channels")
	}

	unlock, err := s.lock.Lock(ctx)
	if err != nil {
		return err
	}
	defer unlock()

	for _, msg := range msgs {
		ch := s.getChannel(msg.ChannelID)
		if ch == nil {
			return xerrors.Errorf("%w", domain.ErrInvalidChannel)
		}
		if ch.log[msg.MessageLocator] != nil {
			continue // Duplicated message
		}

		ch.channelClock = ch.channelClock + 1 // Must start with 1
		wrapped := onmemoryMessage{
			channelClock: ch.channelClock,
			ExpireAt:     domain.Time{Time: s.systemClock.Now().Add(ch.Expire().Duration)},
			Message:      msg,
		}
		if err := wrapped.Validate(); err != nil {
			return err
		}
		ch.log[msg.MessageLocator] = &wrapped

		for _, sbsc := range ch.subscribers {
			sbsc.addMessage(wrapped)
			sbsc.lastActivity = s.systemClock.Now()
		}
	}
	return nil
}

func (s *onmemoryStorage) FetchMessages(ctx context.Context, sl domain.SubscriberLocator, max int, waituntil domain.Duration) (messages []domain.Message, moreMessages bool, ackHandle domain.AckHandle, err error) {
	sbsc, err := s.findSubscriberForFetchMessages(ctx, sl)
	if err != nil {
		return []domain.Message{}, false, domain.AckHandle{}, err
	}

	endPolling := make(chan bool, 1)
	received := make(chan domain.Message, max)
	completed := make(chan error, 2)

	var full int32
	atomic.StoreInt32(&full, 0)
	go func() {
		defer close(received)
		defer func() { completed <- nil }()

		pollingInterval := time.NewTicker(300 * time.Millisecond)
		defer pollingInterval.Stop()

		found := false
	P:
		for {
			func() {
				unlock, err := s.lock.Lock(ctx)
				if err != nil {
					completed <- err
					return
				}
				defer unlock()

				sbsc.lastActivity = s.systemClock.Now()
				// Fetch messages as possible
				for _, msg := range sbsc.messages {
					select {
					case received <- msg.Message: // Receive message
						found = true
					default: // Queue is full (reached to max)
						atomic.StoreInt32(&full, 1)
					}
				}
			}()

			// If message(s) found, return them immediately.
			if found || (atomic.LoadInt32(&full) == 1) {
				break P
			}
			select {
			case <-ctx.Done():
				completed <- ctx.Err()
				return
			case <-endPolling:
				break P
			case <-pollingInterval.C:
				continue P
			}
		}
	}()
	timeoutTimer := time.NewTimer(waituntil.Duration)
	select {
	case err = <-completed: // Completed before timeout
		if err != nil {
			return []domain.Message{}, false, domain.AckHandle{}, err
		}
	case <-timeoutTimer.C:
		endPolling <- true
		if err := <-completed; err != nil {
			return []domain.Message{}, false, domain.AckHandle{}, err
		}
	}

	messages = []domain.Message{}
	for msg := range received {
		messages = append(messages, msg)
	}
	moreMessages = (atomic.LoadInt32(&full) == int32(1))

	if len(messages) > 0 {
		ackHandle = encodeAckHandle(sl, ackHandleData{
			LastMessageID: messages[len(messages)-1].MessageID,
		})
	} else {
		ackHandle = domain.AckHandle{}
	}
	return
}

func (s *onmemoryStorage) AcknowledgeMessages(ctx context.Context, handle domain.AckHandle) error {
	unlock, err := s.lock.Lock(ctx)
	if err != nil {
		return err
	}
	defer unlock()

	ch := s.getChannel(handle.ChannelID)
	if ch == nil {
		return xerrors.Errorf("%w", domain.ErrInvalidChannel)
	}

	sbsc := ch.subscribers[handle.SubscriberID]
	if sbsc == nil {
		return xerrors.Errorf("%w", domain.ErrSubscriptionNotFound)
	}
	sbsc.lastActivity = s.systemClock.Now()

	rhd, err := decodeAckHandle(handle)
	if err != nil {
		return err
	}

	var readUntil = -1
	for i, msg := range sbsc.messages {
		if rhd.LastMessageID == msg.MessageID {
			readUntil = i
			break
		}
	}
	if readUntil == -1 {
		return nil // AckHandle is stale, may be already consumed
	}
	sbsc.channelClock = sbsc.messages[readUntil].channelClock
	sbsc.messages = sbsc.messages[readUntil+1:]
	return nil
}

func (s *onmemoryStorage) IsOldMessages(ctx context.Context, sl domain.SubscriberLocator, msgs []domain.MessageLocator) (map[domain.MessageLocator]bool, error) {
	unlock, err := s.lock.Lock(ctx)
	if err != nil {
		return nil, err
	}
	defer unlock()

	ch := s.getChannel(sl.ChannelID)
	if ch == nil {
		return nil, xerrors.Errorf("%w", domain.ErrInvalidChannel)
	}

	sbsc := ch.subscribers[sl.SubscriberID]
	if sbsc == nil {
		return nil, xerrors.Errorf("%w", domain.ErrSubscriptionNotFound)
	}

	result := map[domain.MessageLocator]bool{}
	for _, msg := range msgs {
		wrapped := ch.log[msg]
		if wrapped != nil && wrapped.channelClock <= sbsc.channelClock {
			result[msg] = true
		} else {
			result[msg] = false
		}
	}
	return result, nil
}
