package channel

import (
	"errors"
	"sync"
	"time"

	"github.com/saiya/dsps/server/domain"
	"golang.org/x/xerrors"
)

const cachedChannelCleanupFactor = 2
const cachedChannelNegativeCacheExpire = 5 * time.Minute

func newCachedChannelProvider(inner domain.ChannelProvider, clock domain.SystemClock) domain.ChannelProvider {
	return &cachedChannels{
		inner:      inner,
		clock:      clock,
		fdPressure: inner.GetFileDescriptorPressure(),

		lock: sync.Mutex{},
		m:    make(map[domain.ChannelID]*cachedChannelEntry, 1024),
		age:  0,
	}
}

type cachedChannels struct {
	inner      domain.ChannelProvider
	clock      domain.SystemClock
	fdPressure int

	// Writer lock blocks all further Rlocks, but reader won't take lock so long time in this usecase.
	lock sync.Mutex
	m    map[domain.ChannelID]*cachedChannelEntry
	age  uint64
}

func (cache *cachedChannels) GetFileDescriptorPressure() int {
	return cache.fdPressure
}

func (cache *cachedChannels) Get(id domain.ChannelID) (domain.Channel, error) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	if ent, ok := cache.m[id]; ok {
		ent.extend(cache.clock)
		if ent.channel == nil {
			return nil, domain.ErrInvalidChannel
		}
		return ent.channel, nil
	}

	cache.cleanup()
	cache.age++

	c, err := cache.inner.Get(id)
	if err != nil && !errors.Is(err, domain.ErrInvalidChannel) {
		return nil, xerrors.Errorf(`channel configuration error on "%s": %w`, id, err)
	}
	ent := &cachedChannelEntry{channel: c}
	ent.extend(cache.clock)
	cache.m[id] = ent
	if c == nil {
		return nil, domain.ErrInvalidChannel
	}
	return c, nil
}

type cachedChannelEntry struct {
	expireAt time.Time
	channel  domain.Channel // nil-able, nil means negative cache.
}

func (entry *cachedChannelEntry) extend(clock domain.SystemClock) {
	if entry.channel != nil {
		entry.expireAt = clock.Now().Add(entry.channel.Expire().Duration)
	} else {
		entry.expireAt = clock.Now().Add(cachedChannelNegativeCacheExpire)
	}
}

func (cache *cachedChannels) cleanup() {
	if cache.age <= uint64(len(cache.m)/cachedChannelCleanupFactor) {
		return
	}

	now := cache.clock.Now()
	for id, entry := range cache.m {
		if entry.expireAt.Before(now.Time) {
			delete(cache.m, id)
		}
	}
	cache.age = 0
}
