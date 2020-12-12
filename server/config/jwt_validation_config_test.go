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
		iss:
			- https://issuer.example.com/issuer-url
		keys:
			RS256:
				- "../jwt/testdata/RS256-2048bit-public.pem"
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
		iss:
			- https://issuer.example.com/issuer-url
		keys:
			none: []
			RS256:
				- "../jwt/testdata/RS256-2048bit-public.pem"
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
	assert.Equal(t, []domain.JwtIss{domain.JwtIss("https://issuer.example.com/issuer-url")}, jwt.Iss)
	assert.Equal(t, []string{}, jwt.Keys["none"])
	assert.Equal(t, []string{"../jwt/testdata/RS256-2048bit-public.pem"}, jwt.Keys["RS256"])
	assert.Equal(t, "{{.regex.id}}", jwt.Claims["chatroom"].String())
}
