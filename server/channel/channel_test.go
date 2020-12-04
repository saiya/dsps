package channel_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/channel"
)

func TestChannelExpire(t *testing.T) {
	assert.Equal(t, 35*time.Minute, channel.NewChannelByAtomYamls(t, "test", []string{
		`{ regex: '.+', expire: '35m' }`,
	}).Expire().Duration)
	assert.Equal(t, 105*time.Minute, channel.NewChannelByAtomYamls(t, "test", []string{
		`{ regex: '.+', expire: '35m' }`,
		`{ regex: '.+', expire: '105m' }`,
	}).Expire().Duration)
}
