package config_test

import (
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
		alg: RS256
		iss:
			- https://issuer.example.com/issuer-url
		keys:
			- "../testing/testdata/RS256-sample-public-key.pem"
`, "\t", "  ")
	config, err := ParseConfig(Overrides{}, configYaml)
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
		alg: RS256
		iss:
			- https://issuer.example.com/issuer-url
		keys:
			- "../testing/testdata/RS256-sample-public-key.pem"
		claims:
			chatroom: '{{.regex.id}}'
`, "\t", "  ")
	config, err := ParseConfig(Overrides{}, configYaml)
	if err != nil {
		t.Error(err)
		return
	}

	cfg := config.Channels[0]
	assert.Equal(t, "chat-room-(?P<id>\\d+)", cfg.Regex.String())

	jwt := cfg.Jwt
	assert.Equal(t, domain.JwtAlg("RS256"), jwt.Alg)
	assert.Equal(t, []domain.JwtIss{domain.JwtIss("https://issuer.example.com/issuer-url")}, jwt.Iss)
	assert.Equal(t, "../testing/testdata/RS256-sample-public-key.pem", jwt.Keys[0])
	assert.Equal(t, "{{.regex.id}}", jwt.Claims["chatroom"].String())
}
