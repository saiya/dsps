package domain

// AckHandle is an token to remove received (acknowledged) messages from a subscriber.
type AckHandle struct {
	SubscriberLocator
	Handle string
}
