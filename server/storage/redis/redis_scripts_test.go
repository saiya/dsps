package redis

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestLoadScripts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s, cmd := newMockedRedisStorage(ctrl)
	cmd.EXPECT().LoadScript(gomock.Any(), gomock.Any()).MinTimes(1)
	assert.NoError(t, s.loadScripts(context.Background()))
}
