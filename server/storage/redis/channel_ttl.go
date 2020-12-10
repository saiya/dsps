package redis

import (
	"math"
	"strconv"

	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/domain"
)

type channelTTLSec int64

// go-redis depends on BinaryMarshaler
func (c channelTTLSec) MarshalBinary() (data []byte, err error) {
	return []byte(strconv.FormatInt(int64(c), 10)), nil
}

func (s *redisStorage) channelRedisTTLSec(channelID domain.ChannelID) (channelTTLSec, error) {
	ch := s.channelProvider(channelID)
	if ch == nil {
		return 0, xerrors.Errorf("%w", domain.ErrInvalidChannel)
	}
	return channelTTLSec(math.Ceil((ch.Expire().Duration + ttlMargin).Seconds())), nil
}
