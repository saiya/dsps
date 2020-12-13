package domain_test

import (
	"testing"
	"time"

	. "github.com/saiya/dsps/server/domain"
	"github.com/stretchr/testify/assert"
)

func TestHoge(t *testing.T) {
	before := time.Now()
	now := RealSystemClock.Now().Time
	after := time.Now()

	assert.True(t, before.Before(now))
	assert.True(t, now.Before(after))
}
