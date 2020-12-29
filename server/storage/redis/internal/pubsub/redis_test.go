package pubsub

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

type redisRawPubSubStub struct {
	t *testing.T
	c chan *redis.Message

	closeOnce       sync.Once
	closed          chan interface{}
	closeResultLock sync.Mutex
	closeResult     error

	pingResultQueue chan error
	messageQueue    chan interface{}

	channelInit  sync.Once
	channelClose sync.Once
	channel      chan *redis.Message
}

func newRedisRawPubSubStub(t *testing.T) *redisRawPubSubStub {
	return &redisRawPubSubStub{
		t:      t,
		c:      make(chan *redis.Message, 1024),
		closed: make(chan interface{}),

		pingResultQueue: make(chan error, 1024),
		messageQueue:    make(chan interface{}, 1024),
	}
}

// Push ping result to queue.
func (s *redisRawPubSubStub) EnqueuePingResult(nCalls int, result error) *redisRawPubSubStub {
	for i := 0; i < nCalls; i++ {
		s.pingResultQueue <- result
	}
	return s
}

func (s *redisRawPubSubStub) EnqueuePingResultForever(result error) *redisRawPubSubStub {
	go func() {
		for {
			select {
			case s.pingResultQueue <- result:
			case <-s.closed:
				return
			}
		}
	}()
	return s
}

func (s *redisRawPubSubStub) Ping(ctx context.Context, payload ...string) error {
	select {
	case <-s.closed:
		return errors.New("stub RedisRawPubSub closed")
	case result := <-s.pingResultQueue:
		return result
	default:
		assert.FailNow(s.t, "Unexpected ping call")
		panic("unreachable code")
	}
}

// Push message to Receive() queue.
// Also able to queue error to cause subscription abort.
func (s *redisRawPubSubStub) EnqueueMessage(msg interface{}) {
	s.messageQueue <- msg
}

func (s *redisRawPubSubStub) EnqueueDefaultSubscribeMessage() *redisRawPubSubStub {
	s.EnqueueMessage(&redis.Subscription{Kind: "psubscribe"})
	return s
}

func (s *redisRawPubSubStub) EnqueueEvent(channel RedisChannelID) {
	s.EnqueueMessage(&redis.Message{Channel: string(channel)})
}

func (s *redisRawPubSubStub) Receive(ctx context.Context) (interface{}, error) {
	return s.receive()
}

func (s *redisRawPubSubStub) receive() (interface{}, error) {
	select {
	case <-s.closed:
		return nil, errors.New("stub RedisRawPubSub closed")
	case msg := <-s.messageQueue:
		switch value := msg.(type) {
		case error:
			return nil, value
		default:
			return msg, nil
		}
	}
}

func (s *redisRawPubSubStub) ChannelSize(size int) <-chan *redis.Message {
	isValidCall := false
	s.channelInit.Do(func() {
		isValidCall = true
		s.channel = make(chan *redis.Message, size)
		go s.receiveWorker()
	})
	assert.True(s.t, isValidCall, "do not call RedisRawPubSub.ChannelSize() twice")
	return s.channel
}

func (s *redisRawPubSubStub) CloseChannel() {
	s.channelClose.Do(func() {
		close(s.channel)
	})
}

func (s *redisRawPubSubStub) receiveWorker() {
	for {
		raw, err := s.receive()
		if err != nil {
			s.t.Logf("background receive worker of RedisRawPubSub stopped because of: %v", err)
			return
		}
		switch value := raw.(type) {
		case *redis.Message:
			select {
			case s.channel <- value:
			default:
				s.t.Logf("background receive worker of RedisRawPubSub could not pass message to channel because closed.")
			}
		default:
			s.t.Logf("background receive worker of RedisRawPubSub dropped message: %v", raw)
		}
	}
}

func (s *redisRawPubSubStub) SetCloseResult(err error) {
	s.closeResultLock.Lock()
	defer s.closeResultLock.Unlock()
	s.closeResult = err
}

func (s *redisRawPubSubStub) Close() error {
	s.closeOnce.Do(func() {
		close(s.closed)
	})

	s.closeResultLock.Lock()
	defer s.closeResultLock.Unlock()
	return s.closeResult
}

func (s *redisRawPubSubStub) IsClosed() bool {
	select {
	case <-s.closed:
		return true
	default:
		return false
	}
}
