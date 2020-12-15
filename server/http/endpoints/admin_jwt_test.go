package endpoints_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/domain"
	. "github.com/saiya/dsps/server/domain/mock"
	. "github.com/saiya/dsps/server/http"
	. "github.com/saiya/dsps/server/http/testing"
)

func TestJwtRevokeWithoutPubSubSupport(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	storage := NewMockStorage(ctrl)
	storage.EXPECT().AsPubSubStorage().Return(nil).AnyTimes()
	storage.EXPECT().AsJwtStorage().Return(nil).AnyTimes()

	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {
		deps.Storage = storage
	}, func(deps *ServerDependencies, baseURL string) {
		res := DoHTTPRequestWithHeaders(t, "PUT", baseURL+"/admin/jwt/revoke?jti=my-jwt&exp=4070912400", AdminAuthHeaders(t, deps), ``)
		AssertErrorResponse(t, res, 501, nil, `No JWT compatible storage available`)
	})
}

func TestJwtRevokeSuccess(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jti := domain.JwtJti("my-jwt")
	exp, err := domain.ParseJwtExp("4070912400")
	assert.NoError(t, err)
	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {}, func(deps *ServerDependencies, baseURL string) {
		revoked, err := deps.Storage.AsJwtStorage().IsRevokedJwt(ctx, jti)
		assert.NoError(t, err)
		assert.False(t, revoked)

		res := DoHTTPRequestWithHeaders(t, "PUT", baseURL+fmt.Sprintf("/admin/jwt/revoke?jti=%s&exp=%s", jti, exp), AdminAuthHeaders(t, deps), ``)
		AssertResponseJSON(t, res, 200, map[string]interface{}{
			"jti": string(jti),
			"exp": float64(exp.Int64()),
		})

		revoked, err = deps.Storage.AsJwtStorage().IsRevokedJwt(ctx, jti)
		assert.NoError(t, err)
		assert.True(t, revoked)

		// Should be idempotent
		res = DoHTTPRequestWithHeaders(t, "PUT", baseURL+fmt.Sprintf("/admin/jwt/revoke?jti=%s&exp=%s", jti, exp), AdminAuthHeaders(t, deps), ``)
		AssertResponseJSON(t, res, 200, map[string]interface{}{
			"jti": string(jti),
			"exp": float64(exp.Int64()),
		})
	})
}

func TestJwtRevokeFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	storage, _, jwt := NewMockStorages(ctrl)

	jti := domain.JwtJti("my-jwt")
	exp, err := domain.ParseJwtExp("4070912400")
	assert.NoError(t, err)
	WithServer(t, `logging: category: "*": FATAL`, func(deps *ServerDependencies) {
		deps.Storage = storage
	}, func(deps *ServerDependencies, baseURL string) {
		res := DoHTTPRequestWithHeaders(t, "PUT", baseURL+fmt.Sprintf("/admin/jwt/revoke?jti=%s&exp=%s", "", exp), AdminAuthHeaders(t, deps), ``)
		AssertErrorResponse(t, res, 400, nil, `Missing "jti" parameter`)

		res = DoHTTPRequestWithHeaders(t, "PUT", baseURL+fmt.Sprintf("/admin/jwt/revoke?jti=%s&exp=%s", jti, "INVALID-EXP"), AdminAuthHeaders(t, deps), ``)
		AssertErrorResponse(t, res, 400, nil, `Invalid "exp" parameter`)

		jwt.EXPECT().RevokeJwt(gomock.Any(), exp, jti).Return(errors.New("mock error"))
		res = DoHTTPRequestWithHeaders(t, "PUT", baseURL+fmt.Sprintf("/admin/jwt/revoke?jti=%s&exp=%s", jti, exp), AdminAuthHeaders(t, deps), ``)
		AssertInternalServerErrorResponse(t, res)
	})
}
