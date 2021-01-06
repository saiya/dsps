package validator_test

import (
	"context"
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	. "github.com/saiya/dsps/server/jwt/testing"
	. "github.com/saiya/dsps/server/jwt/validator"
)

var pregeneratedPublicKeys = map[domain.JwtAlg][]string{
	"RS256": {
		"../testdata/RS256-2048bit-public.pem",
		"../testdata/RS256-4096bit-public.pem",
	},
	"PS256": {
		"../testdata/RS256-2048bit-public.pem",
		"../testdata/RS256-4096bit-public.pem",
	},
	"ES512": {
		"../testdata/ES512-test1-public.pem",
		"../testdata/ES512-test2-public.pem",
	},
	"HS256": {
		"../testdata/HS256.rand",
	},
}

func TestPregeneratedJwt(t *testing.T) {
	var err error
	ctx := context.Background()
	v := createDefaultValidator(t)

	// Valid JWT
	jwt, err := ioutil.ReadFile("../testdata/RS256-2048bit.jwt")
	assert.NoError(t, err)
	assert.NoError(t, v.Validate(ctx, string(jwt)))

	// Expired JWT
	jwt, err = ioutil.ReadFile("../testdata/RS256-2048bit-expired.jwt")
	assert.NoError(t, err)
	err = v.Validate(ctx, string(jwt))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT validation failed: token is expired")
}

func TestAlgorithms(t *testing.T) {
	ctx := context.Background()
	v := createDefaultValidator(t)
	for supported, keyfile := range map[string]string{
		// ref: https://tools.ietf.org/html/rfc7518#section-3.1
		"HS256": "HS256",         // HMAC
		"RS256": "RS256-2048bit", // RSASSA-PKCS1
		"ES512": "ES512-test1",   // ECDSA
		"PS256": "RS256-2048bit", // RSASSA-PSS
	} {
		assert.NoError(t, v.Validate(ctx, GenerateJwt(t, JwtProps{
			Alg:     domain.JwtAlg(supported),
			Keyname: keyfile,
			Iss:     "https://example.com/issuer",
			Aud:     []domain.JwtAud{"https://example.com/audience"},
		})), "alg=%s", supported)
	}
}

func TestMinimalValidation(t *testing.T) {
	ctx := context.Background()
	tpl, err := NewTemplate(ctx, &config.JwtValidationConfig{
		Iss: []domain.JwtIss{"https://example.com/issuer"},
		// No signing
		Keys: map[domain.JwtAlg][]string{"none": {}},
		// No "aud" validation etc.
		ClockSkewLeeway: &domain.Duration{},
	}, domain.RealSystemClock)
	assert.NoError(t, err)
	v, err := tpl.NewValidator(struct{}{})
	assert.NoError(t, err)

	assert.NoError(t, v.Validate(ctx, GenerateJwt(t, JwtProps{
		Alg: "none", Iss: "https://example.com/issuer",
	})))
}

func TestIss(t *testing.T) {
	ctx := context.Background()
	v := createDefaultValidator(t)

	// Valid
	for _, iss := range []domain.JwtIss{"https://example.com/issuer", "https://example.com/issuer2"} {
		assert.NoError(t, v.Validate(ctx, GenerateJwt(t, JwtProps{
			Keyname: "RS256-2048bit", Alg: "RS256",
			Iss: iss,
			Aud: []domain.JwtAud{"https://example.com/audience"},
		})))
	}

	// Invalid
	for _, iss := range []domain.JwtIss{"https://example.com/issuer3", ""} {
		err := v.Validate(ctx, GenerateJwt(t, JwtProps{
			Keyname: "RS256-2048bit", Alg: "RS256",
			Iss: iss,
			Aud: []domain.JwtAud{"https://example.com/audience"},
		}))
		assert.Error(t, err)
		assert.Regexp(t, `"iss" claim of the presented JWT .+ does not match with any of expected values`, err.Error())
	}
}

func TestAud(t *testing.T) {
	ctx := context.Background()
	v := createDefaultValidator(t)

	// Valid
	for _, audList := range [][]domain.JwtAud{
		{"https://example.com/audience"},
		{"https://example.com/audience2"},
		{"https://example.com/audience", "https://example.com/audience2"},
		{"D0DE4D1E-C175-44CE-96E7-29528B9D8040", "https://example.com/audience2"},
	} {
		assert.NoError(t, v.Validate(ctx, GenerateJwt(t, JwtProps{
			Keyname: "RS256-2048bit", Alg: "RS256",
			Iss: "https://example.com/issuer",
			Aud: audList,
		})))
	}

	// Invalid
	for _, testcase := range []struct {
		Error   string
		AudList []domain.JwtAud
	}{
		{Error: `no "aud" claim found on the presented JWT`},
		{Error: `"aud" claim of the presented JWT .+ does not match with any of expected values`, AudList: []domain.JwtAud{"https://example.com/audience3"}},
		{Error: `"aud" claim of the presented JWT .+ does not match with any of expected values`, AudList: []domain.JwtAud{"D0DE4D1E-C175-44CE-96E7-29528B9D8040", "55065412-63B1-49FF-9C86-3E82435857AE"}},
	} {
		err := v.Validate(ctx, GenerateJwt(t, JwtProps{
			Keyname: "RS256-2048bit", Alg: "RS256",
			Iss: "https://example.com/issuer",
			Aud: testcase.AudList,
		}))
		assert.Error(t, err)
		assert.Regexp(t, testcase.Error, err.Error())
	}
}

func TestMultipleKeyMatching(t *testing.T) {
	ctx := context.Background()
	v := createDefaultValidator(t)

	// Validator must try all possible keys
	for _, props := range []JwtProps{
		{Keyname: "ES512-test1", Alg: "ES512"},
		{Keyname: "ES512-test2", Alg: "ES512"},
		{Keyname: "RS256-2048bit", Alg: "RS256"},
		{Keyname: "RS256-4096bit", Alg: "RS256"},
	} {
		props.Iss = "https://example.com/issuer"
		props.Aud = []domain.JwtAud{"https://example.com/audience"}
		assert.NoError(t, v.Validate(ctx, GenerateJwt(t, props)))
	}
}

func TestValidDuration(t *testing.T) {
	var err error
	ctx := context.Background()
	v := createDefaultValidator(t)

	// Valid (present minimal claims)
	assert.NoError(t, v.Validate(ctx, GenerateJwt(t, JwtProps{
		Keyname: "ES512-test1",
		Alg:     "ES512",
		Iss:     "https://example.com/issuer",
		Aud:     []domain.JwtAud{"https://example.com/audience"},
	})))

	// Valid (present all claims)
	assert.NoError(t, v.Validate(ctx, GenerateJwt(t, JwtProps{
		Keyname: "ES512-test1",
		Alg:     "ES512",
		Iss:     "https://example.com/issuer",
		Aud:     []domain.JwtAud{"https://example.com/audience"},

		Nbf: time.Now().Add(-1 * time.Second),
		Iat: time.Now().Add(-1 * time.Second),
		Exp: time.Now().Add(+15 * time.Second),
	})))

	// Invalid
	for _, testcase := range []struct {
		Message                string
		NbfSec, IatSec, ExpSec time.Duration
	}{
		{
			Message: `"iat" claim value of the presented JWT is in future`,
			NbfSec:  -1, IatSec: +10, ExpSec: +15,
		},
		{
			Message: `token is not valid yet`,
			NbfSec:  +10, IatSec: +10, ExpSec: +15,
		},
		{
			Message: `token is expired`,
			NbfSec:  -20, IatSec: -15, ExpSec: -10,
		},
	} {
		err = v.Validate(ctx, GenerateJwt(t, JwtProps{
			Keyname: "ES512-test1",
			Alg:     "ES512",
			Iss:     "https://example.com/issuer",
			Aud:     []domain.JwtAud{"https://example.com/audience"},

			Nbf: time.Now().Add(testcase.NbfSec * time.Second),
			Iat: time.Now().Add(testcase.IatSec * time.Second),
			Exp: time.Now().Add(testcase.ExpSec * time.Second),
		}))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), testcase.Message)
	}
}

func TestClockLeeway(t *testing.T) {
	ctx := context.Background()
	tpl, err := NewTemplate(ctx, &config.JwtValidationConfig{
		Iss:             []domain.JwtIss{"https://example.com/issuer"},
		Keys:            pregeneratedPublicKeys,
		ClockSkewLeeway: &domain.Duration{Duration: 300 * time.Second},
	}, domain.RealSystemClock)
	assert.NoError(t, err)
	v, err := tpl.NewValidator(struct{}{})
	assert.NoError(t, err)

	assert.Equal(t, domain.Duration{Duration: 300 * time.Second}, tpl.JWTClockSkewLeewayMax())

	// Valid, within clock skew leeway
	for _, testcase := range []struct {
		NbfSec, IatSec, ExpSec time.Duration
	}{
		{NbfSec: -1, IatSec: +10, ExpSec: +15},
		{NbfSec: +10, IatSec: +10, ExpSec: +15},
		{NbfSec: -20, IatSec: -15, ExpSec: -10},
	} {
		assert.NoError(t, v.Validate(ctx, GenerateJwt(t, JwtProps{
			Keyname: "ES512-test1",
			Alg:     "ES512",
			Iss:     "https://example.com/issuer",

			Nbf: time.Now().Add(testcase.NbfSec * time.Second),
			Iat: time.Now().Add(testcase.IatSec * time.Second),
			Exp: time.Now().Add(testcase.ExpSec * time.Second),
		})))
	}

	// Invalid, out of range
	for _, testcase := range []struct {
		Message                string
		NbfSec, IatSec, ExpSec time.Duration
	}{
		{
			Message: `"iat" claim value of the presented JWT is in future`,
			NbfSec:  -1, IatSec: +10 + 600, ExpSec: +15 + 600,
		},
		{
			Message: `token is not valid yet`,
			NbfSec:  +10 + 600, IatSec: +10 + 600, ExpSec: +15 + 600,
		},
		{
			Message: `token is expired`,
			NbfSec:  -20 - 600, IatSec: -15 - 600, ExpSec: -10 - 600,
		},
	} {
		err = v.Validate(ctx, GenerateJwt(t, JwtProps{
			Keyname: "ES512-test1",
			Alg:     "ES512",
			Iss:     "https://example.com/issuer",

			Nbf: time.Now().Add(testcase.NbfSec * time.Second),
			Iat: time.Now().Add(testcase.IatSec * time.Second),
			Exp: time.Now().Add(testcase.ExpSec * time.Second),
		}))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), testcase.Message)
	}
}

func TestCustomClaims(t *testing.T) {
	chatroomTpl, err := domain.NewTemplateString(`{{.channel.id}}`)
	assert.NoError(t, err)
	asterTpl, err := domain.NewTemplateString(`*`)
	assert.NoError(t, err)
	trueTpl, err := domain.NewTemplateString(`true`)
	assert.NoError(t, err)

	ctx := context.Background()
	tpl, err := NewTemplate(ctx, &config.JwtValidationConfig{
		Iss:  []domain.JwtIss{"https://example.com/issuer"},
		Keys: pregeneratedPublicKeys,
		Claims: map[string]domain.TemplateStrings{
			"chatroom":                       domain.NewTemplateStrings(chatroomTpl, asterTpl),
			"https://example.com/payed-user": domain.NewTemplateStrings(trueTpl),
		},
		ClockSkewLeeway: &domain.Duration{},
	}, domain.RealSystemClock)
	assert.NoError(t, err)
	v, err := tpl.NewValidator(map[string]map[string]string{
		"channel": {
			"id": "1234",
		},
	})
	assert.NoError(t, err)

	// Valid
	for _, claims := range []map[string]interface{}{
		{"chatroom": "1234", "https://example.com/payed-user": "true"},
		{"chatroom": "*", "https://example.com/payed-user": "true"},
		{"chatroom": 1234, "https://example.com/payed-user": true},
	} {
		assert.NoError(t, v.Validate(ctx, GenerateJwt(t, JwtProps{
			Keyname: "ES512-test1",
			Alg:     "ES512",
			Iss:     "https://example.com/issuer",
			Claims:  claims,
		})))
	}

	// Invalid
	for _, testcase := range []struct {
		Message string
		Claims  map[string]interface{}
	}{
		{
			Message: `required ".+" claim by setting but not present or non-string value presented in the JWT`,
			Claims:  map[string]interface{}{},
		},
		{
			Message: `required "https://example.com/payed-user" claim by setting but not present or non-string value presented in the JWT`,
			Claims:  map[string]interface{}{"chatroom": "1234"},
		},
		{
			Message: `required "chatroom" claim to be \[1234 \*\] by setting but presented JWT has value "9999"`,
			Claims:  map[string]interface{}{"chatroom": 9999, "https://example.com/payed-user": true},
		},
		{
			Message: `required "https://example.com/payed-user" claim to be \[true\] by setting but presented JWT has value "INVALID"`,
			Claims:  map[string]interface{}{"chatroom": 1234, "https://example.com/payed-user": "INVALID"},
		},
	} {
		err := v.Validate(ctx, GenerateJwt(t, JwtProps{
			Keyname: "ES512-test1",
			Alg:     "ES512",
			Iss:     "https://example.com/issuer",
			Claims:  testcase.Claims,
		}))
		assert.Error(t, err)
		assert.Regexp(t, testcase.Message, err.Error())
	}
}

func createDefaultValidator(t *testing.T) Validator {
	ctx := context.Background()
	tpl, err := NewTemplate(ctx, &config.JwtValidationConfig{
		Iss:             []domain.JwtIss{"https://example.com/issuer", "https://example.com/issuer2"},
		Aud:             []domain.JwtAud{"https://example.com/audience", "https://example.com/audience2"},
		Keys:            pregeneratedPublicKeys,
		ClockSkewLeeway: &domain.Duration{},
	}, domain.RealSystemClock)
	assert.NoError(t, err)
	v, err := tpl.NewValidator(struct{}{})
	assert.NoError(t, err)
	return v
}
