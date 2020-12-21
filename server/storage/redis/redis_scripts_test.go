package redis

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

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

func TestScriptLoader(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s, cmd := newMockedRedisStorage(ctrl)
	loadScriptCalledOnce := sync.Once{}
	loadScriptCalled := make(chan interface{})
	cmd.EXPECT().LoadScript(gomock.Any(), gomock.Any()).MinTimes(1).DoAndReturn(func(context.Context, interface{}) error {
		loadScriptCalledOnce.Do(func() {
			close(loadScriptCalled)
		})
		return errors.New("test error")
	})

	loader := s.startScriptLoader(context.Background(), 1*time.Millisecond)
	defer loader.stopScriptLoader(context.Background())

	<-loadScriptCalled
	time.Sleep(50 * time.Millisecond)
}
