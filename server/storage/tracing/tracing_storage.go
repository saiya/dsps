package tracing

import "github.com/saiya/dsps/server/domain"

// NewTracingStorage wraps given Storage to trace calls
func NewTracingStorage(s domain.Storage) (domain.Storage, error) {
	return s, nil // TODO: Implement
}

/*
type tracingStorageMethodCallStat struct {
	Liveness  int64 `json:"liveness"`
	Readiness int64 `json:"readiness"`

	Stat             int64 `json:"stat"`
	NewSubscriber    int64 `json:"newSubscriber"`
	RemoveSubscriber int64 `json:"deleteSubscriber"`

	PublishMessages     int64 `json:"publishMessages"`
	FetchMessages       int64 `json:"fetchMessages"`
	AcknowledgeMessages int64 `json:"acknowledgeMessages"`
	IsOldMessages       int64 `json:"isOldMessages"`

	RevokeJwt    int64 `json:"revokeJwt"`
	IsRevokedJwt int64 `json:"isRevokedJwt"`
}
*/
