package channel

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/domain"
	dspstesting "github.com/saiya/dsps/server/testing"
)

func TestCacheExpiry(t *testing.T) {
	clock := dspstesting.NewStubClock(t)
	p := newCachedChannelProvider(func(id domain.ChannelID) domain.Channel {
		return NewChannelByAtomYamls(t, id, []string{`{ regex: ".+", expire: "1s" }`})
	}, clock)

	test1 := p("test1")
	assert.NotNil(t, test1)
	assert.Same(t, test1, p("test1"))
	assert.Same(t, test1, p("test1"))

	// Enforce GC
	clock.Add(2 * time.Second)
	test2 := p("test2")
	assert.NotNil(t, test2)
	assert.NotSame(t, test1, p("test1")) // GC collected
	test1 = p("test1")
	assert.Same(t, test1, p("test1"))

	// Enforce GC
	for i := 0; i < 10; i++ {
		clock.Add(500 * time.Millisecond)
		assert.NotNil(t, p("test2")) // Keep touching to test2
		p(domain.ChannelID(fmt.Sprintf("test2-add-age-%d", i)))
	}
	assert.Same(t, test2, p("test2"))    // Still alive
	assert.NotSame(t, test1, p("test1")) // GC collected
}

func TestNegativeCache(t *testing.T) {
	clock := dspstesting.NewStubClock(t)
	notFoundCount := 0
	p := newCachedChannelProvider(func(id domain.ChannelID) domain.Channel {
		if strings.HasPrefix(string(id), "not-found-") {
			notFoundCount++
			return nil
		}
		return NewChannelByAtomYamls(t, id, []string{`{ regex: ".+", expire: "1s" }`})
	}, clock)

	assert.Nil(t, p("not-found-zero"))
	assert.Equal(t, 1, notFoundCount)
	assert.Nil(t, p("not-found-zero"))
	assert.Equal(t, 1, notFoundCount) // Cached

	// Enforce GC
	clock.Add(cachedChannelNegativeCacheExpire + 1*time.Microsecond)
	assert.NotNil(t, p("test-1"))
	assert.NotNil(t, p("test-2"))

	assert.Nil(t, p("not-found-zero"))
	assert.Equal(t, 2, notFoundCount) // Cache evicted
}
