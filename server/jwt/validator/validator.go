package validator

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	jwtgo "github.com/dgrijalva/jwt-go/v4"

	"github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	"github.com/saiya/dsps/server/jwt"
)

// Template is a template of Validator
type Template interface {
	NewValidator(tplEnv domain.TemplateStringEnv) (Validator, error)
}

type validatorTemplate struct {
	cfg   *config.JwtValidationConfig
	clock domain.SystemClock

	validAlgs []string
	keysMap   map[domain.JwtAlg][]interface{}
	parser    *jwtgo.Parser
}

// Validator is a object to validate JWT
type Validator interface {
	Validate(ctx context.Context, jwt string) error
}

type validator struct {
	validatorTemplate

	claims map[string]string
}

// NewTemplate creates Template instance.
func NewTemplate(ctx context.Context, cfg *config.JwtValidationConfig, clock domain.SystemClock) (Template, error) {
	validAlgs := make([]string, 0, len(cfg.Keys))
	keysMap := make(map[domain.JwtAlg][]interface{}, len(cfg.Keys))
	for alg, keyFilesOrg := range cfg.Keys {
		keyFiles := make([]string, len(keyFilesOrg))
		copy(keyFiles, keyFilesOrg)
		if alg.IsNone() {
			keyFiles = append(keyFiles, "dummy") // Ensure magic key value added
		}

		validAlgs = append(validAlgs, string(alg))
		keys := make([]interface{}, 0, len(keyFiles))
		for _, kf := range keyFiles {
			key, err := jwt.LoadVerificationKey(alg, kf)
			if err != nil {
				return nil, err
			}
			keys = append(keys, key)
		}
		keysMap[alg] = keys
	}

	return &validatorTemplate{
		cfg:   cfg,
		clock: clock,

		validAlgs: validAlgs,
		keysMap:   keysMap,
		parser: jwtgo.NewParser(
			jwtgo.WithValidMethods(validAlgs),
			jwtgo.WithLeeway(cfg.ClockSkewLeeway.Duration),
			jwtgo.WithoutAudienceValidation(), // Because jwt-go does not support multiple candidate values
		),
	}, nil
}

func (v *validatorTemplate) NewValidator(tplEnv domain.TemplateStringEnv) (Validator, error) {
	claims := make(map[string]string, len(v.cfg.Claims))
	for claim, tpl := range v.cfg.Claims {
		str, err := tpl.Execute(tplEnv)
		if err != nil {
			return nil, fmt.Errorf(`failed to evaluate template string of JWT "%s" claim configuration "%s": %w`, claim, tpl, err)
		}
		claims[claim] = str
	}
	return &validator{validatorTemplate: *v, claims: claims}, nil
}

func (v *validator) Validate(ctx context.Context, jwt string) error {
	if jwt == "" {
		return fmt.Errorf("no JWT presented")
	}

	// Parse and validate "exp", "nbf", "iat", signature
	claims := jwtgo.MapClaims{}
	var keyFuncError error = nil
	if _, err := v.parser.ParseWithClaims(jwt, claims, func(t *jwtgo.Token) (interface{}, error) {
		key, err := v.findKeyCandidate(t, jwt)
		keyFuncError = err
		return key, err
	}); err != nil {
		if keyFuncError != nil {
			return fmt.Errorf("JWT validation failed: %w", keyFuncError)
		}
		return fmt.Errorf("JWT validation failed: %w", err)
	}

	if err := v.validateIat(ctx, claims); err != nil { // Validate "iat"
		return err
	}
	if err := v.validateIss(ctx, claims); err != nil { // Validate "iss"
		return err
	}
	if err := v.validateAud(ctx, claims); err != nil { // Validate "aud"
		return err
	}
	if err := v.validateCustomClaims(ctx, claims); err != nil { // Validate user-defined claims
		return err
	}
	return nil
}

func (v *validator) validateIat(ctx context.Context, claims jwtgo.MapClaims) error {
	value, err := claims.LoadTimeValue("iat")
	if err != nil {
		return fmt.Errorf(`failed to parse "iat" claim value: %w`, err)
	}
	if value == nil {
		return nil
	}
	if v.parser.ValidationHelper.Before(value.Time) {
		return fmt.Errorf(`"iat" claim value of the presented JWT is in future`)
	}
	return nil
}

func (v *validator) validateIss(ctx context.Context, claims jwtgo.MapClaims) error {
	for _, acceptable := range v.cfg.Iss {
		err := claims.VerifyIssuer(v.parser.ValidationHelper, string(acceptable))
		if err == nil {
			return nil // One matches successfully
		}
	}
	return fmt.Errorf(`"iss" claim of the presented JWT ("%v") does not match with any of expected values (%v)`, claims["iss"], v.cfg.Iss)
}

func (v *validator) validateAud(ctx context.Context, claims jwtgo.MapClaims) error {
	if len(v.cfg.Aud) == 0 { // "aud" validation not configured
		return nil
	}
	actual, err := jwtgo.ParseClaimStrings(claims["aud"])
	if err != nil {
		return fmt.Errorf(`failed to parse "aud" claim of the presented JWT: %w`, err)
	}
	if actual == nil {
		return fmt.Errorf(`no "aud" claim found on the presented JWT`)
	}

	for _, acceptable := range v.cfg.Aud {
		// requires go-jwt >= 4.0 because older versions are not RFC-compliant https://github.com/dgrijalva/jwt-go/issues/348
		err := v.parser.ValidationHelper.ValidateAudienceAgainst(actual, string(acceptable))
		if err == nil {
			return nil // One matches successfully
		}
	}
	return fmt.Errorf(`"aud" claim of the presented JWT ("%v") does not match with any of expected values (%v)`, actual, v.cfg.Aud)
}

func (v *validator) validateCustomClaims(ctx context.Context, claims jwtgo.MapClaims) error {
	for claim, expected := range v.claims {
		var value string
		switch raw := claims[claim].(type) {
		case string:
			value = raw
		case float64:
			value = strconv.FormatFloat(raw, 'f', -1, 64)
		case bool:
			value = fmt.Sprintf("%t", raw)
		default:
			return fmt.Errorf(`required "%s" claim by setting but not present or non-string value presented in the JWT`, claim)
		}
		if value != expected {
			return fmt.Errorf(`required "%s" claim to be "%s" by setting but presented JWT has value "%s"`, claim, expected, value)
		}
	}
	return nil
}

func (v *validator) findKeyCandidate(t *jwtgo.Token, jwt string) (interface{}, error) {
	alg := domain.JwtAlg(t.Method.Alg())
	keys := v.keysMap[alg]
	switch len(keys) {
	case 0:
		return nil, fmt.Errorf(`signing algorithm "%s" of the presented JWT is not in configured allow list %v`, alg, v.validAlgs)
	case 1:
		return keys[0], nil
	default:
		// Currently go-jwt does not support matching multiple keys.
		// https://github.com/dgrijalva/jwt-go/issues/416
		token, parts, err := v.parser.ParseUnverified(jwt, jwtgo.MapClaims{})
		if err != nil {
			return keys[0], err
		}
		for _, key := range keys {
			if err = token.Method.Verify(strings.Join(parts[0:2], "."), parts[2], key); err != nil {
				continue
			}
			return key, nil
		}
		return nil, fmt.Errorf(`no matching %s signing key found for the presented JWT: %w`, alg, err)
	}
}
