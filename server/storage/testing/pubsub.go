package testing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/domain"
	dspstesting "github.com/saiya/dsps/server/testing"
)

// PubSubTest tests common Storage behaviors
func PubSubTest(t *testing.T, storageCtor StorageCtor) {
	storageSubTest(t, storageCtor, "pubSubScenario", _pubSubScenarioTest)
	storageSubTest(t, storageCtor, "longPolling", _longPollingTest)
	storageSubTest(t, storageCtor, "longPollngCancel", _longPollingCancelTest)
	storageSubTest(t, storageCtor, "messageAgeDetection", _messageAgeDetectionTest)
	storageSubTest(t, storageCtor, "manyQueuesMessages", _manyQueuesMessagesTest)
	storageSubTest(t, storageCtor, "pubSubMixedChannePublish", _pubSubMixedChannelPublishTest)
	storageSubTest(t, storageCtor, "pubSubInvalidChannel", _pubSubInvalidChannelTest)
	storageSubTest(t, storageCtor, "pubsubInvalidSubscriber", _pubsubInvalidSubscriber)
	storageSubTest(t, storageCtor, "pubSubInvalidMessage", _pubSubInvalidMessageTest)
}

func _pubSubScenarioTest(t *testing.T, storageCtor StorageCtor) {
	ctx := context.Background()
	s, err := storageCtor(ctx, domain.RealSystemClock, StubChannelProvider)
	if !assert.NoError(t, err) {
		return
	}
	defer func() { assert.NoError(t, s.Shutdown(ctx)) }()
	storage := s.AsPubSubStorage()
	assert.NotNil(t, storage)

	var ch domain.ChannelID = randomChannelID()
	sl := domain.SubscriberLocator{
		ChannelID:    ch,
		SubscriberID: "sbsc1",
	}

	// Create subscriber
	if !assert.NoError(t, storage.NewSubscriber(ctx, sl)) {
		return
	}
	if !assert.NoError(t, storage.NewSubscriber(ctx, sl)) { // Must be idempotent
		return
	}
	defer func() { assert.NoError(t, storage.RemoveSubscriber(ctx, sl)) }()
	defer func() { assert.NoError(t, storage.RemoveSubscriber(ctx, sl)) }() // Must be idempotent

	// Empty short&long polling
	if received, more, _, err := storage.FetchMessages(ctx, sl, 10, dspstesting.MakeDuration("0ms")); assert.NoError(t, err) {
		assert.Equal(t, 0, len(received))
		assert.False(t, more)
	}
	if received, _, _, err := storage.FetchMessages(ctx, sl, 10, dspstesting.MakeDuration("100ms")); assert.NoError(t, err) {
		assert.Equal(t, 0, len(received))
	}

	// Publish messages
	var messages = make([]domain.Message, 16)
	for i := range messages {
		messages[i] = domain.Message{
			MessageLocator: domain.MessageLocator{
				ChannelID: ch,
				MessageID: domain.MessageID(fmt.Sprintf("msg-%d", i)),
			},
			Content: []byte(fmt.Sprintf("{\"hi\":\"hello %d\"}", i)),
		}
	}
	assert.NoError(t, storage.PublishMessages(ctx, []domain.Message{})) // Publish 0 messages (no-op)
	if !assert.NoError(t, storage.PublishMessages(ctx, messages)) {
		return
	}

	// Receive messages (with small bulkSize)
	received, more, _, err := storage.FetchMessages(ctx, sl, len(messages)-1, dspstesting.MakeDuration("0ms"))
	if !assert.NoError(t, err) {
		return
	}
	dspstesting.MessagesEqual(t, messages[0:len(messages)-1], received)
	assert.True(t, more) // More messages

	// Receive messages again (with larger bulkSize)
	receivedAgain, more, ackHandle, err := storage.FetchMessages(ctx, sl, len(messages)+1, dspstesting.MakeDuration("0ms"))
	if !assert.NoError(t, err) {
		return
	}
	dspstesting.MessagesEqual(t, messages, receivedAgain)
	assert.False(t, more)

	// Remove messages from the subscriber
	if !assert.NoError(t, storage.AcknowledgeMessages(ctx, ackHandle)) {
		return
	}
	if !assert.NoError(t, storage.AcknowledgeMessages(ctx, ackHandle)) { // Must be idempotent
		return
	}

	// Receive after deletion
	receivedAfterDelete, more, _, err := storage.FetchMessages(ctx, sl, len(messages), dspstesting.MakeDuration("0ms"))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, 0, len(receivedAfterDelete))
	assert.False(t, more)

	// Publish duplicated messages again (should be ignored)
	if !assert.NoError(t, storage.PublishMessages(ctx, messages)) {
		return
	}
	receivedAfterResend, more, _, err := storage.FetchMessages(ctx, sl, len(messages), dspstesting.MakeDuration("0ms"))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, 0, len(receivedAfterResend))
	assert.False(t, more)

	// Publish another message
	moreMessages := []domain.Message{
		{
			MessageLocator: domain.MessageLocator{
				ChannelID: ch,
				MessageID: "additional-msg",
			},
			Content: []byte("{\"hi\":\"hello, again\"}"),
		},
	}
	if !assert.NoError(t, storage.PublishMessages(ctx, moreMessages)) {
		return
	}
	receivedMoreMessages, more, _, err := storage.FetchMessages(ctx, sl, len(messages), dspstesting.MakeDuration("0ms"))
	if !assert.NoError(t, err) {
		return
	}
	dspstesting.MessagesEqual(t, moreMessages, receivedMoreMessages)
	assert.False(t, more)

	// Acknowledge stale AckHandle, should have NO side-effects.
	if !assert.NoError(t, storage.AcknowledgeMessages(ctx, ackHandle)) {
		return
	}
	receivedMoreMessagesAgain, more, _, err := storage.FetchMessages(ctx, sl, len(messages), dspstesting.MakeDuration("0ms"))
	if !assert.NoError(t, err) {
		return
	}
	dspstesting.MessagesEqual(t, moreMessages, receivedMoreMessagesAgain)
	assert.False(t, more)
}

func _manyQueuesMessagesTest(t *testing.T, storageCtor StorageCtor) {
	ctx := context.Background()
	s, err := storageCtor(ctx, domain.RealSystemClock, StubChannelProvider)
	if !assert.NoError(t, err) {
		return
	}
	defer func() { assert.NoError(t, s.Shutdown(ctx)) }()
	storage := s.AsPubSubStorage()
	assert.NotNil(t, storage)

	var ch domain.ChannelID = randomChannelID()
	sl := domain.SubscriberLocator{
		ChannelID:    ch,
		SubscriberID: "sbsc1",
	}

	// Create subscriber
	if !assert.NoError(t, storage.NewSubscriber(ctx, sl)) {
		return
	}
	defer func() { assert.NoError(t, storage.RemoveSubscriber(ctx, sl)) }()

	// Publish messages
	var messages = make([]domain.Message, 32)
	for i := range messages {
		messages[i] = domain.Message{
			MessageLocator: domain.MessageLocator{
				ChannelID: ch,
				MessageID: domain.MessageID(fmt.Sprintf("msg-%d", i)),
			},
			Content: []byte(fmt.Sprintf("{\"hi\":\"hello %d\"}", i)),
		}
	}
	if !assert.NoError(t, storage.PublishMessages(ctx, messages)) {
		return
	}

	// Receive messages partially
	received1, more, rcvHandle1, err := storage.FetchMessages(ctx, sl, 16, dspstesting.MakeDuration("0ms"))
	if !assert.NoError(t, err) {
		return
	}
	assert.True(t, more)
	dspstesting.MessagesEqual(t, messages[0:16], received1)
	if !assert.NoError(t, storage.AcknowledgeMessages(ctx, rcvHandle1)) {
		return
	}

	// Receive messages remaining
	received2, more, rcvHandle2, err := storage.FetchMessages(ctx, sl, 16, dspstesting.MakeDuration("0ms"))
	if !assert.NoError(t, err) {
		return
	}
	assert.False(t, more)
	dspstesting.MessagesEqual(t, messages[16:32], received2)
	if !assert.NoError(t, storage.AcknowledgeMessages(ctx, rcvHandle2)) {
		return
	}
}

func _longPollingTest(t *testing.T, storageCtor StorageCtor) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second) // Prevent test blocks
	defer ctxCancel()
	s, err := storageCtor(ctx, domain.RealSystemClock, StubChannelProvider)
	if !assert.NoError(t, err) {
		return
	}
	defer func() { assert.NoError(t, s.Shutdown(ctx)) }()
	storage := s.AsPubSubStorage()
	assert.NotNil(t, storage)

	var ch domain.ChannelID = randomChannelID()
	sl := domain.SubscriberLocator{
		ChannelID:    ch,
		SubscriberID: "sbsc1",
	}

	// Create subscriber
	if !assert.NoError(t, storage.NewSubscriber(ctx, sl)) {
		return
	}
	defer func() { assert.NoError(t, storage.RemoveSubscriber(ctx, sl)) }()

	// Publish messages asynchronously
	var messages = make([]domain.Message, 16)
	for i := range messages {
		messages[i] = domain.Message{
			MessageLocator: domain.MessageLocator{
				ChannelID: ch,
				MessageID: domain.MessageID(fmt.Sprintf("msg-%d", i)),
			},
			Content: []byte(fmt.Sprintf("{\"hi\":\"hello %d\"}", i)),
		}
	}

	willPublish := make(chan interface{})
	publishedAt := make(chan time.Time, 1)
	go func() {
		<-willPublish
		time.Sleep(dspstesting.MakeDuration("1ms").Duration)
		assert.NoError(t, storage.PublishMessages(ctx, messages))
		publishedAt <- time.Now()
	}()

	// short polling (run before publish)
	receivedBeforePublish, more, _, _ := storage.FetchMessages(ctx, sl, len(messages), dspstesting.MakeDuration("0ms"))
	assert.Equal(t, 0, len(receivedBeforePublish))
	assert.False(t, more)

	// long polling
	close(willPublish)
	received, more, _, err := storage.FetchMessages(ctx, sl, len(messages), dspstesting.MakeDuration("120s")) // Must return immediately after publish
	if assert.NoError(t, err) {
		receivedAt := time.Now()
		dspstesting.MessagesEqual(t, messages, received)
		assert.False(t, more)
		assert.LessOrEqual(t, receivedAt.Sub(<-publishedAt).Seconds(), (3 * time.Second).Seconds())
	}
}

func _longPollingCancelTest(t *testing.T, storageCtor StorageCtor) {
	ctx := context.Background()
	s, err := storageCtor(ctx, domain.RealSystemClock, StubChannelProvider)
	if !assert.NoError(t, err) {
		return
	}
	defer func() { assert.NoError(t, s.Shutdown(ctx)) }()
	storage := s.AsPubSubStorage()
	assert.NotNil(t, storage)

	var ch domain.ChannelID = randomChannelID()
	sl := domain.SubscriberLocator{
		ChannelID:    ch,
		SubscriberID: "sbsc1",
	}

	// Create subscriber
	if !assert.NoError(t, storage.NewSubscriber(ctx, sl)) {
		return
	}
	defer func() { assert.NoError(t, storage.RemoveSubscriber(ctx, sl)) }()

	cancelableCtx, cancel := context.WithCancel(ctx)
	beforeCancel := time.Now()
	go func() {
		time.Sleep(dspstesting.MakeDuration("100ms").Duration)
		cancel()
	}()
	_, _, _, err = storage.FetchMessages(cancelableCtx, sl, 1, dspstesting.MakeDuration("12s"))
	afterEnd := time.Now()
	assert.Less(t, afterEnd.Sub(beforeCancel).Milliseconds(), int64(300))
	dspstesting.IsError(t, context.Canceled, err)
}

func _messageAgeDetectionTest(t *testing.T, storageCtor StorageCtor) {
	ctx := context.Background()
	s, err := storageCtor(ctx, domain.RealSystemClock, StubChannelProvider)
	if !assert.NoError(t, err) {
		return
	}
	defer func() { assert.NoError(t, s.Shutdown(ctx)) }()
	storage := s.AsPubSubStorage()
	assert.NotNil(t, storage)

	var ch domain.ChannelID = randomChannelID()
	var messages = make([]domain.Message, 5)
	var locators = make([]domain.MessageLocator, len(messages))
	for i := range messages {
		messages[i] = domain.Message{
			MessageLocator: domain.MessageLocator{
				ChannelID: ch,
				MessageID: domain.MessageID(fmt.Sprintf("msg-%d", i)),
			},
			Content: []byte(fmt.Sprintf("{\"hi\":\"hello %d\"}", i)),
		}
		locators[i] = messages[i].MessageLocator
	}

	// msg[0] : Unsent (lost) message
	// msg[1] : Sent before subscriber creation
	assert.NoError(t, storage.PublishMessages(ctx, messages[1:2]))

	sl := domain.SubscriberLocator{
		ChannelID:    ch,
		SubscriberID: "sbsc1",
	}
	if !assert.NoError(t, storage.NewSubscriber(ctx, sl)) {
		return
	}
	defer func() { assert.NoError(t, storage.RemoveSubscriber(ctx, sl)) }()

	ageMap, err := storage.IsOldMessages(ctx, sl, locators)
	assert.NoError(t, err)
	assert.Equal(t, map[domain.MessageLocator]bool{
		locators[0]: false, // Unsent message is/may not older than subscriber
		locators[1]: true,  // Sent before subscriber
		locators[2]: false, // Not yet sent
		locators[3]: false, // Not yet sent
		locators[4]: false, // Not yet sent
	}, ageMap)

	// Sent msg[2] after subscriber creation
	assert.NoError(t, storage.PublishMessages(ctx, messages[2:3]))

	// Fetch msg[2]
	received, more, ackHandle, err := storage.FetchMessages(ctx, sl, 1, dspstesting.MakeDuration("0s"))
	assert.NoError(t, err)
	assert.False(t, more)
	assert.Equal(t, messages[2:3], received)
	assert.Equal(t, map[domain.MessageLocator]bool{
		locators[0]: false, // Unsent message is/may not older than subscriber
		locators[1]: true,  // Sent before subscriber
		locators[2]: false, // Fetched but not yet acknowledged
		locators[3]: false, // Not yet sent
		locators[4]: false, // Not yet sent
	}, ageMap)

	// Ack msg[2]
	assert.NoError(t, storage.AcknowledgeMessages(ctx, ackHandle))
	ageMap, err = storage.IsOldMessages(ctx, sl, locators)
	assert.NoError(t, err)
	assert.Equal(t, map[domain.MessageLocator]bool{
		locators[0]: false, // Unsent message is/may not older than subscriber
		locators[1]: true,  // Sent before subscriber
		locators[2]: true,  // Acknowledged
		locators[3]: false, // Not yet sent
		locators[4]: false, // Not yet sent
	}, ageMap)

	// Send msg[3] also.
	assert.NoError(t, storage.PublishMessages(ctx, messages[3:4]))
	ageMap, err = storage.IsOldMessages(ctx, sl, locators)
	assert.NoError(t, err)
	assert.Equal(t, map[domain.MessageLocator]bool{
		locators[0]: false, // Unsent message is/may not older than subscriber
		locators[1]: true,  // Sent before subscriber
		locators[2]: true,  // Acknowledged
		locators[3]: false, // Not yet fetched nor acknowledged
		locators[4]: false, // Not yet sent
	}, ageMap)

	// Fetch msg[3]
	received, more, ackHandle, err = storage.FetchMessages(ctx, sl, 1, dspstesting.MakeDuration("0s"))
	assert.NoError(t, err)
	assert.False(t, more)
	assert.Equal(t, messages[3:4], received)

	// Send msg[4] (before Ack of msg[3])
	assert.NoError(t, storage.PublishMessages(ctx, messages[4:5]))

	// Fetch msg[4] (before Ack of msg[3])
	received, more, _, err = storage.FetchMessages(ctx, sl, 2, dspstesting.MakeDuration("0s"))
	assert.NoError(t, err)
	assert.False(t, more)
	assert.Equal(t, messages[3:5], received)

	// Ack msg[3], but not msg[4]
	assert.NoError(t, storage.AcknowledgeMessages(ctx, ackHandle))
	ageMap, err = storage.IsOldMessages(ctx, sl, locators)
	assert.NoError(t, err)
	assert.Equal(t, map[domain.MessageLocator]bool{
		locators[0]: false, // Unsent message is/may not older than subscriber
		locators[1]: true,  // Sent before subscriber
		locators[2]: true,  // Acknowledged
		locators[3]: true,  // Acknowledged
		locators[4]: false, // Fetched but not yet acknowledged
	}, ageMap)
}

func _pubSubMixedChannelPublishTest(t *testing.T, storageCtor StorageCtor) {
	ctx := context.Background()
	s, err := storageCtor(ctx, domain.RealSystemClock, StubChannelProvider)
	if !assert.NoError(t, err) {
		return
	}
	defer func() { assert.NoError(t, s.Shutdown(ctx)) }()
	storage := s.AsPubSubStorage()
	assert.NotNil(t, storage)

	assert.Error(t, storage.PublishMessages(ctx, []domain.Message{
		{
			MessageLocator: domain.MessageLocator{
				ChannelID: randomChannelID(),
				MessageID: "msg1",
			},
			Content: []byte(`{}`),
		},
		{
			MessageLocator: domain.MessageLocator{
				ChannelID: randomChannelID(),
				MessageID: "msg2",
			},
			Content: []byte(`{}`),
		},
	}))
}

func _pubSubInvalidChannelTest(t *testing.T, storageCtor StorageCtor) {
	ctx := context.Background()
	s, err := storageCtor(ctx, domain.RealSystemClock, StubChannelProvider)
	if !assert.NoError(t, err) {
		return
	}
	defer func() { assert.NoError(t, s.Shutdown(ctx)) }()
	storage := s.AsPubSubStorage()
	assert.NotNil(t, storage)

	validCh := randomChannelID()
	validSL := domain.SubscriberLocator{
		ChannelID:    validCh,
		SubscriberID: "sbsc1",
	}
	assert.NoError(t, storage.NewSubscriber(ctx, validSL))
	assert.NoError(t, storage.PublishMessages(ctx, []domain.Message{
		{
			MessageLocator: domain.MessageLocator{
				ChannelID: validCh,
				MessageID: "msg-to-valid-channel",
			},
			Content: []byte("{ \"hi\": \"hello\" }"),
		},
	}))
	_, _, ackHandle, err := storage.FetchMessages(ctx, validSL, 1, dspstesting.MakeDuration("0s"))
	assert.NoError(t, err)

	ch := DisabledChannelID
	sl := domain.SubscriberLocator{
		ChannelID:    DisabledChannelID,
		SubscriberID: "sbsc2",
	}

	dspstesting.IsError(t, domain.ErrInvalidChannel, storage.NewSubscriber(ctx, sl))
	dspstesting.IsError(t, domain.ErrInvalidChannel, storage.PublishMessages(ctx, []domain.Message{
		{
			MessageLocator: domain.MessageLocator{
				ChannelID: ch,
				MessageID: "msg-to-invalid-channel",
			},
			Content: []byte("{\"hi\":\"hello, again\"}"),
		},
	}))
	if _, _, _, err := storage.FetchMessages(ctx, sl, 1, dspstesting.MakeDuration("0s")); !dspstesting.IsOneOfErrors(t, []error{domain.ErrInvalidChannel, domain.ErrSubscriptionNotFound}, err) {
		return
	}
	dspstesting.IsError(t, domain.ErrInvalidChannel, storage.AcknowledgeMessages(ctx, domain.AckHandle{
		SubscriberLocator: sl,
		Handle:            ackHandle.Handle,
	}))

	// IsOldMessages can return false for invalid channel/receiver, or able to return error
	isOldTarget := domain.MessageLocator{
		ChannelID: ch,
		MessageID: "msg-to-undefined-subscriber",
	}
	isOld, err := storage.IsOldMessages(ctx, sl, []domain.MessageLocator{isOldTarget})
	assert.True(t, (!isOld[isOldTarget]) || errors.Is(err, domain.ErrSubscriptionNotFound))

	// RemoveSubscriber should return success for non-existent subscriber/channel
	assert.NoError(t, storage.RemoveSubscriber(ctx, domain.SubscriberLocator{
		ChannelID:    ch,
		SubscriberID: "invalid-channel-recv",
	}))
}

func _pubsubInvalidSubscriber(t *testing.T, storageCtor StorageCtor) {
	ctx := context.Background()
	s, err := storageCtor(ctx, domain.RealSystemClock, StubChannelProvider)
	if !assert.NoError(t, err) {
		return
	}
	defer func() { assert.NoError(t, s.Shutdown(ctx)) }()
	storage := s.AsPubSubStorage()
	assert.NotNil(t, storage)

	ch := randomChannelID()
	validSL := domain.SubscriberLocator{
		ChannelID:    ch,
		SubscriberID: "sbsc1",
	}
	assert.NoError(t, storage.NewSubscriber(ctx, validSL))
	assert.NoError(t, storage.PublishMessages(ctx, []domain.Message{
		{
			MessageLocator: domain.MessageLocator{
				ChannelID: ch,
				MessageID: "msg-to-valid-channel",
			},
			Content: []byte("{ \"hi\": \"hello\" }"),
		},
	}))
	_, _, ackHandle, err := storage.FetchMessages(ctx, validSL, 1, dspstesting.MakeDuration("0s"))
	assert.NoError(t, err)

	sl := domain.SubscriberLocator{
		ChannelID:    ch,
		SubscriberID: "undefined-subscriber",
	}
	_, _, _, err = storage.FetchMessages(ctx, sl, 1, dspstesting.MakeDuration("10s"))
	dspstesting.IsError(t, domain.ErrSubscriptionNotFound, err)
	dspstesting.IsOneOfErrors(t, []error{domain.ErrSubscriptionNotFound, domain.ErrMalformedAckHandle}, storage.AcknowledgeMessages(ctx, domain.AckHandle{
		SubscriberLocator: sl,
		Handle:            ackHandle.Handle,
	}))

	// IsOldMessages can return false for invalid channel/receiver, or able to return error
	isOldTarget := domain.MessageLocator{
		ChannelID: ch,
		MessageID: "msg-to-undefined-subscriber",
	}
	isOld, err := storage.IsOldMessages(ctx, sl, []domain.MessageLocator{isOldTarget})
	assert.True(t, (!isOld[isOldTarget]) || errors.Is(err, domain.ErrSubscriptionNotFound))
}

func randomChannelID() domain.ChannelID {
	uuid, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	return domain.ChannelID(fmt.Sprintf("ch-%s", uuid))
}

func _pubSubInvalidMessageTest(t *testing.T, storageCtor StorageCtor) {
	ctx := context.Background()
	s, err := storageCtor(ctx, domain.RealSystemClock, StubChannelProvider)
	if !assert.NoError(t, err) {
		return
	}
	defer func() { assert.NoError(t, s.Shutdown(ctx)) }()
	storage := s.AsPubSubStorage()
	assert.NotNil(t, storage)

	ch := randomChannelID()
	dspstesting.IsError(t, domain.ErrMalformedMessageJSON, storage.PublishMessages(ctx, []domain.Message{
		{
			MessageLocator: domain.MessageLocator{
				ChannelID: ch,
				MessageID: "msg-1",
			},
			Content: json.RawMessage(`INVALID JSON`),
		},
	}))
}
