package redis

import (
	"math"
	"strconv"
	"time"

	"github.com/saiya/dsps/server/domain"
)

type channelTTLSec int64

// go-redis depends on BinaryMarshaler
func (c channelTTLSec) MarshalBinary() (data []byte, err error) {
	return []byte(strconv.FormatInt(int64(c), 10)), nil
}

func (c channelTTLSec) asDuration() time.Duration {
	return time.Duration(c) * time.Second
}

func (s *redisStorage) channelRedisTTLSec(channelID domain.ChannelID) (channelTTLSec, error) {
	ch, err := s.channelProvider.Get(channelID)
	if err != nil {
		return 0, err
	}
	return channelTTLSec(math.Ceil((ch.Expire().Duration + ttlMargin).Seconds())), nil
}
