package redis

import (
	"context"

	"github.com/saiya/dsps/server/domain"
)

func (s *redisStorage) RevokeJwt(ctx context.Context, exp domain.JwtExp, jti domain.JwtJti) error {
	d := exp.Time().Sub(s.clock.Now().Time) + ttlMargin
	if d <= 0 {
		return nil
	}
	return s.RedisCmd.SetEX(ctx, keyOfJti(jti).Revocation(), exp.String(), d)
}

func (s *redisStorage) IsRevokedJwt(ctx context.Context, jti domain.JwtJti) (bool, error) {
	value, err := s.RedisCmd.Get(ctx, keyOfJti(jti).Revocation())
	if err != nil {
		return false, err
	}
	if value == nil {
		return false, nil
	}
	return true, nil
}
