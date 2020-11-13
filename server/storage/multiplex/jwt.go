package multiplex

import (
	"context"

	"github.com/saiya/dsps/server/domain"
)

func (s *storageMultiplexer) RevokeJwt(ctx context.Context, exp domain.JwtExp, jti domain.JwtJti) error {
	_, err := s.parallelAtLeastOneSuccess(ctx, "RevokeJwt", func(ctx context.Context, _ domain.StorageID, child domain.Storage) (interface{}, error) {
		if child := child.AsJwtStorage(); child != nil {
			return nil, child.RevokeJwt(ctx, exp, jti)
		}
		return nil, errMultiplexSkipped
	})
	return err
}

func (s *storageMultiplexer) IsRevokedJwt(ctx context.Context, jti domain.JwtJti) (bool, error) {
	results, err := s.parallelAtLeastOneSuccess(ctx, "IsRevokedJwt", func(ctx context.Context, _ domain.StorageID, child domain.Storage) (interface{}, error) {
		if child := child.AsJwtStorage(); child != nil {
			return child.IsRevokedJwt(ctx, jti)
		}
		return nil, errMultiplexSkipped
	})
	if err != nil {
		return false, err
	}
	for _, result := range results {
		if result.(bool) {
			return true, nil
		}
	}
	return false, nil
}
