package tracing

import (
	"context"

	"github.com/saiya/dsps/server/domain"
)

func (ts *tracingStorage) RevokeJwt(ctx context.Context, exp domain.JwtExp, jti domain.JwtJti) error {
	return ts.jwt.RevokeJwt(ctx, exp, jti)
}

func (ts *tracingStorage) IsRevokedJwt(ctx context.Context, jti domain.JwtJti) (bool, error) {
	return ts.jwt.IsRevokedJwt(ctx, jti)
}
