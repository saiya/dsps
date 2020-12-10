package redis

import "strconv"

func parseChannelClock(value string) *channelClock {
	i := parseRedisInt64(value)
	if i == nil {
		return nil
	}

	result := channelClock(*i)
	return &result
}

func parseRedisInt64(value string) *int64 {
	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil
	}
	return &result
}
