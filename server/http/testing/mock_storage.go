package testing

import (
	"github.com/golang/mock/gomock"

	"github.com/saiya/dsps/server/domain/mock"
)

// NewMockStorages creates storage mock objects
func NewMockStorages(ctrl *gomock.Controller) (storage *mock.MockStorage, pubsub *mock.MockPubSubStorage, jwts *mock.MockJwtStorage) {
	storage = mock.NewMockStorage(ctrl)
	pubsub = mock.NewMockPubSubStorage(ctrl)
	jwts = mock.NewMockJwtStorage(ctrl)

	storage.EXPECT().AsPubSubStorage().Return(pubsub).AnyTimes()
	storage.EXPECT().AsJwtStorage().Return(jwts).AnyTimes()
	return
}
