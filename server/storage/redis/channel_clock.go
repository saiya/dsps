package redis

import (
	"strconv"
)

type channelClock int64

// See ../doc/storage/redis-internal-structure.md for why
const clockMin = channelClock(-((int64(1) << 53) - 1))
const clockMax = channelClock((int64(1) << 53) - 1)

func iterateClocks(count int, fromExclusive channelClock, toInclusive channelClock) []channelClock {
	var toInclusive64 int64
	if toInclusive < fromExclusive {
		toInclusive64 = int64(toInclusive) - int64(clockMin) + 1 + int64(clockMax)
	} else {
		toInclusive64 = int64(toInclusive)
	}

	result := make([]channelClock, 0, count)
	for i := int64(1); i <= int64(count); i++ {
		clock := int64(fromExclusive) + i
		if clock > toInclusive64 {
			break
		}
		if clock > int64(clockMax) {
			clock = clock - int64(clockMax) - 1 + int64(clockMin)
		}
		result = append(result, channelClock(clock))
	}
	return result
}

func isClockWithin(clock channelClock, fromExclusive channelClock, toInclusive channelClock) bool {
	if toInclusive < fromExclusive {
		// Range is (from, clockMax] and [clockMin, to]
		return fromExclusive < clock || clock <= toInclusive
	}
	// Range is (from, to]
	return fromExclusive < clock && clock <= toInclusive
}

// go-redis depends on BinaryMarshaler
func (c channelClock) MarshalBinary() (data []byte, err error) {
	return []byte(strconv.FormatInt(int64(c), 10)), nil
}
