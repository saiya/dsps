package channel

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	dspstesting "github.com/saiya/dsps/server/testing"
)

func TestProvider(t *testing.T) {
	cfg, err := config.ParseConfig(context.Background(), config.Overrides{}, `channels: [ { regex: "test.+", expire: "1s" } ]`)
	assert.NoError(t, err)
	clock := dspstesting.NewStubClock(t)
	cp, err := NewChannelProvider(context.Background(), &cfg, clock)
	assert.NoError(t, err)

	test1, err := cp("test1")
	assert.NoError(t, err)
	test1Again, err := cp("test1")
	assert.NoError(t, err)
	assert.NotNil(t, test1)
	assert.Same(t, test1, test1Again)

	notfound, err := cp("not-found")
	assert.Nil(t, notfound)
	dspstesting.IsError(t, domain.ErrInvalidChannel, err)
}
