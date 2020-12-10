package onmemory

import (
	"context"

	"github.com/saiya/dsps/server/domain"
)

func (s *onmemoryStorage) RevokeJwt(ctx context.Context, exp domain.JwtExp, jti domain.JwtJti) error {
	unlock, err := s.lock.Lock(ctx)
	if err != nil {
		return err
	}
	defer unlock()

	s.revokedJwts[jti] = exp
	return nil
}

func (s *onmemoryStorage) IsRevokedJwt(ctx context.Context, jti domain.JwtJti) (bool, error) {
	unlock, err := s.lock.Lock(ctx)
	if err != nil {
		return false, err
	}
	defer unlock()

	exp, found := s.revokedJwts[jti]
	return found && !s.systemClock.Now().After(exp.Time()), nil
}
