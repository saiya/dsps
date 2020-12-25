package middleware_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/domain"
	. "github.com/saiya/dsps/server/http"
	. "github.com/saiya/dsps/server/http/middleware"
	"github.com/saiya/dsps/server/http/router"
	. "github.com/saiya/dsps/server/http/testing"
	. "github.com/saiya/dsps/server/jwt/testing"
	"github.com/saiya/dsps/server/sentry"
)

const jwtDir = "../../jwt"
const configRequiresJWT = `
logging: category: "*": ERROR
channels:
	-
		regex: 'auth-test-channel'
		jwt:
			iss: [ "https://issuer.example.com/issuer-url" ]
			aud: [ "https://my-service.example.com/" ]
			keys:
				RS256: [ "../../jwt/testdata/RS256-2048bit-public.pem" ]
`

func TestNormalAuthFilter(t *testing.T) {
	WithServerDeps(t, configRequiresJWT, func(deps *ServerDependencies) {
		auth := NewNormalAuth(context.Background(), deps, func(context.Context, router.MiddlewareArgs) (Channel, error) {
			return deps.ChannelProvider.Get("auth-test-channel")
		})("", "")

		rec := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Add("Authorization", "Bearer "+GenerateJwt(t, JwtProps{
			Alg:     "RS256",
			Keyname: "RS256-2048bit",
			JwtDir:  jwtDir,
			Iss:     "https://issuer.example.com/issuer-url",
			Aud:     []JwtAud{"https://my-service.example.com/"},
		}))
		withNextFunc(t, true, func(next func(context.Context, router.MiddlewareArgs)) {
			auth(context.Background(), router.MiddlewareArgs{
				HandlerArgs: router.HandlerArgs{R: router.Request{Request: req}, W: router.NewResponseWriter(rec), PS: httprouter.Params{}},
			}, next)
		})
		assert.Equal(t, 200, rec.Code)
	})

	WithServer(t, configRequiresJWT, func(deps *ServerDependencies) {}, func(deps *ServerDependencies, baseURL string) {
		putURL := fmt.Sprintf("%s/channel/%s/message/%s", baseURL, "auth-test-channel", "msg-1")
		assert.NoError(t, deps.Storage.AsJwtStorage().RevokeJwt(context.Background(), JwtExp(time.Now().Add(3*time.Hour)), "revoked-92BFB148-43D0-43B7-9DBE-C7006B4DFB13"))

		// Without JWT
		res := DoHTTPRequest(t, "PUT", putURL, `{}`)
		assert.NoError(t, res.Body.Close())
		assert.Equal(t, 403, res.StatusCode)

		// With JWT (without "jti" claim)
		req, err := http.NewRequestWithContext(context.Background(), "PUT", putURL, strings.NewReader(`{}`))
		assert.NoError(t, err)
		req.Header.Add("Authorization", "Bearer "+GenerateJwt(t, JwtProps{
			Alg:     "RS256",
			Keyname: "RS256-2048bit",
			JwtDir:  jwtDir,
			Iss:     "https://issuer.example.com/issuer-url",
			Aud:     []JwtAud{"https://my-service.example.com/"},
		}))
		res, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, res.StatusCode)

		// With JWT (with "jti" claim)
		req, err = http.NewRequestWithContext(context.Background(), "PUT", putURL, strings.NewReader(`{}`))
		assert.NoError(t, err)
		req.Header.Add("Authorization", "Bearer "+GenerateJwt(t, JwtProps{
			Alg:     "RS256",
			Keyname: "RS256-2048bit",
			JwtDir:  jwtDir,
			Jti:     "FBA4742B-252A-4F28-9834-C181F503D314",
			Iss:     "https://issuer.example.com/issuer-url",
			Aud:     []JwtAud{"https://my-service.example.com/"},
		}))
		res, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, res.StatusCode)

		// With JWT (with revoked "jti")
		req, err = http.NewRequestWithContext(context.Background(), "PUT", putURL, strings.NewReader(`{}`))
		assert.NoError(t, err)
		req.Header.Add("Authorization", "Bearer "+GenerateJwt(t, JwtProps{
			Alg:     "RS256",
			Keyname: "RS256-2048bit",
			JwtDir:  jwtDir,
			Jti:     "revoked-92BFB148-43D0-43B7-9DBE-C7006B4DFB13",
			Iss:     "https://issuer.example.com/issuer-url",
			Aud:     []JwtAud{"https://my-service.example.com/"},
		}))
		res, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, 403, res.StatusCode)
	})
}

func TestNormalAuthSentry(t *testing.T) {
	sentry := sentry.NewStubSentry()
	WithServer(t, configRequiresJWT, func(deps *ServerDependencies) {
		deps.Sentry = sentry
	}, func(deps *ServerDependencies, baseURL string) {
		putURL := fmt.Sprintf("%s/channel/%s/message/%s", baseURL, "auth-test-channel", "msg-1")
		assert.NoError(t, deps.Storage.AsJwtStorage().RevokeJwt(context.Background(), JwtExp(time.Now().Add(3*time.Hour)), "revoked-92BFB148-43D0-43B7-9DBE-C7006B4DFB13"))

		// With JWT (without "jti" claim, invalid issuer)
		req, err := http.NewRequestWithContext(context.Background(), "PUT", putURL, strings.NewReader(`{}`))
		assert.NoError(t, err)
		req.Header.Add("Authorization", "Bearer "+GenerateJwt(t, JwtProps{
			Alg:     "RS256",
			Keyname: "RS256-2048bit",
			JwtDir:  jwtDir,
			Iss:     "https://invalid-issuer.example.com/",
			Aud:     []JwtAud{"https://my-service.example.com/"},
		}))
		res, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, 403, res.StatusCode)

		// With JWT (with revoked "jti")
		req, err = http.NewRequestWithContext(context.Background(), "PUT", putURL, strings.NewReader(`{}`))
		assert.NoError(t, err)
		req.Header.Add("Authorization", "Bearer "+GenerateJwt(t, JwtProps{
			Alg:     "RS256",
			Keyname: "RS256-2048bit",
			JwtDir:  jwtDir,
			Jti:     "revoked-92BFB148-43D0-43B7-9DBE-C7006B4DFB13",
			Iss:     "https://issuer.example.com/issuer-url",
			Aud:     []JwtAud{"https://my-service.example.com/"},
		}))
		res, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, 403, res.StatusCode)
	})

	breadcrumbs := sentry.GetBreadcrumbs()
	assert.Equal(t, 2, len(breadcrumbs))
	issUnmatch := breadcrumbs[0]
	assert.Equal(t, "auth", issUnmatch.Category)
	assert.Regexp(t, `JWT verification failure.+"iss" claim of the presented JWT.+does not match with any of expected values`, issUnmatch.Message)
	revoked := breadcrumbs[1]
	assert.Equal(t, "auth", revoked.Category)
	assert.Regexp(t, `JWT verification failure: presented JWT has been revoked`, revoked.Message)
	assert.Equal(t, "revoked-92BFB148-43D0-43B7-9DBE-C7006B4DFB13", sentry.GetTags()["jti"])
}

func TestNormalAuthFilterPassThrough(t *testing.T) {
	// Without JWT validation configuration, must pass any request
	WithServer(t, ``, func(deps *ServerDependencies) {}, func(deps *ServerDependencies, baseURL string) {
		putURL := fmt.Sprintf("%s/channel/%s/message/%s", baseURL, "auth-test-channel", "msg-1")

		// Without Authorization header
		res := DoHTTPRequest(t, "PUT", putURL, `{}`)
		assert.NoError(t, res.Body.Close())
		assert.Equal(t, 200, res.StatusCode)
	})
}

func TestNormalAuthInvalidChannel(t *testing.T) {
	WithServerDeps(t, configRequiresJWT, func(deps *ServerDependencies) {
		auth := NewNormalAuth(context.Background(), deps, func(context.Context, router.MiddlewareArgs) (Channel, error) {
			return deps.ChannelProvider.Get("INVALID-channel") // Invalid channel ID
		})("", "")

		rec := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Add("Authorization", "Bearer NOT-JWT")
		withNextFunc(t, false, func(next func(context.Context, router.MiddlewareArgs)) {
			auth(context.Background(), router.MiddlewareArgs{HandlerArgs: router.HandlerArgs{R: router.Request{Request: req}, W: router.NewResponseWriter(rec), PS: httprouter.Params{}}}, next)
		})
		AssertRecordedCode(t, rec, http.StatusBadRequest, ErrInvalidChannel)
	})
}

func TestNormalAuthMissingHeader(t *testing.T) {
	WithServerDeps(t, configRequiresJWT+`http: discloseAuthRejectionDetail: true`, func(deps *ServerDependencies) {
		auth := NewNormalAuth(context.Background(), deps, func(context.Context, router.MiddlewareArgs) (Channel, error) {
			return deps.ChannelProvider.Get("auth-test-channel")
		})("", "")

		rec := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil) // No Authorization header
		withNextFunc(t, false, func(next func(context.Context, router.MiddlewareArgs)) {
			auth(context.Background(), router.MiddlewareArgs{HandlerArgs: router.HandlerArgs{R: router.Request{Request: req}, W: router.NewResponseWriter(rec), PS: httprouter.Params{}}}, next)
		})
		AssertRecordedCode(t, rec, http.StatusForbidden, ErrAuthRejection)
		assert.Equal(t, `JWT verification failure: no JWT presented`, BodyJSONMapOfRec(t, rec)["reason"])
	})
}

func TestNormalAuthRejection(t *testing.T) {
	WithServerDeps(t, configRequiresJWT, func(deps *ServerDependencies) {
		auth := NewNormalAuth(context.Background(), deps, func(context.Context, router.MiddlewareArgs) (Channel, error) {
			return deps.ChannelProvider.Get("auth-test-channel")
		})("", "")

		rec := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Add("Authorization", "Bearer NOT-JWT")
		withNextFunc(t, false, func(next func(context.Context, router.MiddlewareArgs)) {
			auth(context.Background(), router.MiddlewareArgs{HandlerArgs: router.HandlerArgs{R: router.Request{Request: req}, W: router.NewResponseWriter(rec), PS: httprouter.Params{}}}, next)
		})
		AssertRecordedCode(t, rec, http.StatusForbidden, ErrAuthRejection)
		assert.Nil(t, BodyJSONMapOfRec(t, rec)["reason"]) // Should not contain detailed message by default
	})
}

func TestNormalAuthRejectionWithDetail(t *testing.T) {
	WithServerDeps(t, configRequiresJWT+`http: discloseAuthRejectionDetail: true`, func(deps *ServerDependencies) {
		auth := NewNormalAuth(context.Background(), deps, func(context.Context, router.MiddlewareArgs) (Channel, error) {
			return deps.ChannelProvider.Get("auth-test-channel")
		})("", "")

		rec := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Add("Authorization", "Bearer NOT-JWT")
		withNextFunc(t, false, func(next func(context.Context, router.MiddlewareArgs)) {
			auth(context.Background(), router.MiddlewareArgs{HandlerArgs: router.HandlerArgs{R: router.Request{Request: req}, W: router.NewResponseWriter(rec), PS: httprouter.Params{}}}, next)
		})
		AssertRecordedCode(t, rec, http.StatusForbidden, ErrAuthRejection)
		assert.Regexp(t, `JWT verification failure.+token is malformed`, BodyJSONMapOfRec(t, rec)["reason"])
	})
}
