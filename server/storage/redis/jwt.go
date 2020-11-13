package redis

import (
	"context"
	"errors"

	"github.com/saiya/dsps/server/domain"
)

func (s *redisStorage) RevokeJwt(ctx context.Context, exp domain.JwtExp, jti domain.JwtJti) error {
	return errors.New("Not Implemented yet")
}

func (s *redisStorage) IsRevokedJwt(ctx context.Context, jti domain.JwtJti) (bool, error) {
	return false, errors.New("Not Implemented yet")
}
