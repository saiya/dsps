package domain

import (
	"context"
	"errors"
)

// StorageID is unique & persistent ID of the Storage
type StorageID string

var (
	// ErrInvalidChannel : Given channel is not permitted in configuration
	ErrInvalidChannel = NewErrorWithCode("dsps.storage.invalid-channel")
	// ErrSubscriptionNotFound : Subscriber expired (due to infrequent access less than expire setting) or not created
	ErrSubscriptionNotFound = NewErrorWithCode("dsps.storage.subscription-not-found")
	// ErrMalformedAckHandle : Failed to decode/decrypt given AckHandle
	ErrMalformedAckHandle = NewErrorWithCode("dsps.storage.ack-handle-malformed")
)

// IsStorageNonFatalError returns true if given error does not indicate storage system error
func IsStorageNonFatalError(err error) bool {
	return errors.Is(err, ErrInvalidChannel) || errors.Is(err, ErrSubscriptionNotFound) || errors.Is(err, ErrMalformedAckHandle)
}

// Storage interface is an abstraction layer of storage implementations
type Storage interface {
	String() string // Stringer
	Shutdown(ctx context.Context) error

	// Liveness probe, returns "encoding/json" encodable value.
	Liveness(ctx context.Context) (interface{}, error)
	// Readiness probe, returns "encoding/json" encodable value.
	Readiness(ctx context.Context) (interface{}, error)

	// Returns statistics, returns "encoding/json" encodable value.
	Stat(ctx context.Context) (interface{}, error)

	// Retruns nil if neither supported nor supported.
	AsPubSubStorage() PubSubStorage
	// Retruns nil if neither supported nor supported.
	AsJwtStorage() JwtStorage
}

// PubSubStorage interface is an abstraction layer of PubSub storage implementations
type PubSubStorage interface {
	NewSubscriber(ctx context.Context, sl SubscriberLocator) error
	RemoveSubscriber(ctx context.Context, sl SubscriberLocator) error

	PublishMessages(ctx context.Context, msgs []Message) error
	// Storage implementation can return more messages than given max count.
	// When length of the returned messages is zero, returned AckHandle is not valid thus caller should ignore it.
	FetchMessages(ctx context.Context, sl SubscriberLocator, max int, waituntil Duration) (messages []Message, moreMessages bool, ackHandle AckHandle, err error)
	AcknowledgeMessages(ctx context.Context, handle AckHandle) error
	// If the message had been acknowledged or sent before subscriber creation, returns true. Otherwise false (can includes unsure messages).
	IsOldMessages(ctx context.Context, sl SubscriberLocator, msgs []MessageLocator) (map[MessageLocator]bool, error)
}

// JwtStorage interface is an abstraction layer of JWT storage implementations
type JwtStorage interface {
	RevokeJwt(ctx context.Context, exp JwtExp, jti JwtJti) error
	IsRevokedJwt(ctx context.Context, jti JwtJti) (bool, error)
}
