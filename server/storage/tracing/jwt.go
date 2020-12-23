package tracing

import (
	"context"

	"github.com/saiya/dsps/server/domain"
)

func (ts *tracingStorage) RevokeJwt(ctx context.Context, exp domain.JwtExp, jti domain.JwtJti) error {
	ctx, end := ts.t.StartStorageSpan(ctx, ts.id, "RevokeJwt")
	ts.t.SetJTI(ctx, jti)
	defer end()
	return ts.jwt.RevokeJwt(ctx, exp, jti)
}

func (ts *tracingStorage) IsRevokedJwt(ctx context.Context, jti domain.JwtJti) (bool, error) {
	ctx, end := ts.t.StartStorageSpan(ctx, ts.id, "IsRevokedJwt")
	ts.t.SetJTI(ctx, jti)
	defer end()
	return ts.jwt.IsRevokedJwt(ctx, jti)
}
