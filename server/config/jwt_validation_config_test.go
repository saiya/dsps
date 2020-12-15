package config_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
	. "github.com/saiya/dsps/server/testing"
)

func TestChannelJwtDefaultConfig(t *testing.T) {
	configYaml := strings.ReplaceAll(`
channels:
-
	regex: 'chat-room-(?P<id>\d+)'
	jwt:
		iss:
			- https://issuer.example.com/issuer-url
		keys:
			RS256:
				- "../jwt/testdata/RS256-2048bit-public.pem"
`, "\t", "  ")
	config, err := ParseConfig(context.Background(), Overrides{}, configYaml)
	if err != nil {
		t.Error(err)
		return
	}

	cfg := config.Channels[0]
	assert.Equal(t, "chat-room-(?P<id>\\d+)", cfg.Regex.String())

	jwt := cfg.Jwt
	assert.Equal(t, MakeDurationPtr("5m"), jwt.ClockSkewLeeway)
	assert.Equal(t, 0, len(jwt.Aud))
	assert.Equal(t, 0, len(jwt.Claims))
}

func TestJwtFullConfig(t *testing.T) {
	configYaml := strings.ReplaceAll(`
channels:
-
	regex: 'chat-room-(?P<id>\d+)'
	jwt:
		iss:
			- https://issuer.example.com/issuer-url
		keys:
			none: []
			RS256:
				- "../jwt/testdata/RS256-2048bit-public.pem"
		claims:
			chatroom: '{{.regex.id}}'
`, "\t", "  ")
	config, err := ParseConfig(context.Background(), Overrides{}, configYaml)
	if err != nil {
		t.Error(err)
		return
	}

	cfg := config.Channels[0]
	assert.Equal(t, "chat-room-(?P<id>\\d+)", cfg.Regex.String())

	jwt := cfg.Jwt
	assert.Equal(t, []domain.JwtIss{domain.JwtIss("https://issuer.example.com/issuer-url")}, jwt.Iss)
	assert.Equal(t, []string{}, jwt.Keys["none"])
	assert.Equal(t, []string{"../jwt/testdata/RS256-2048bit-public.pem"}, jwt.Keys["RS256"])
	assert.Equal(t, "{{.regex.id}}", jwt.Claims["chatroom"].String())
}

func TestJwtConfigError(t *testing.T) {
	_, err := ParseConfig(context.Background(), Overrides{}, `channels: [ { regex: '.+', jwt: { keys: { none: [] } } } ]`)
	assert.Regexp(t, `must supply one or more "iss" \(issuer claim\) list`, err.Error())

	_, err = ParseConfig(context.Background(), Overrides{}, `channels: [ { regex: '.+', jwt: { iss: [ "issuer1" ] } } ]`)
	assert.Regexp(t, `must supply one or more "keys" \(signing algorithm and keys\) setting`, err.Error())

	_, err = ParseConfig(context.Background(), Overrides{}, `channels: [ { regex: '.+', jwt: { iss: [ "issuer1" ], keys: { INVALID: [ "../jwt/testdata/RS256-2048bit-public.pem" ] } } } ]`)
	assert.Regexp(t, `invalid signing algorithm name given "INVALID"`, err.Error())

	_, err = ParseConfig(context.Background(), Overrides{}, `channels: [ { regex: '.+', jwt: { iss: [ "issuer1" ], keys: { RS256: [] } } } ]`)
	assert.Regexp(t, `must supply one or more key file\(s\) to validate JWT signature for alg=RS256`, err.Error())

	_, err = ParseConfig(context.Background(), Overrides{}, `channels: [ { regex: '.+', jwt: { iss: [ "issuer1" ], keys: { RS256: [ "/file/not/found" ] } } } ]`)
	assert.Regexp(t, `failed to read JWT key file "/file/not/found"`, err.Error())
}
