package config_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/config"
	. "github.com/saiya/dsps/server/testing"
)

func TestChannelDefaultConfig(t *testing.T) {
	configYaml := strings.ReplaceAll(`
channels:
-
	regex: 'chat-room-(?P<id>\d+)'
`, "\t", "  ")
	config, err := ParseConfig(Overrides{}, configYaml)
	if err != nil {
		t.Error(err)
		return
	}

	cfg := config.Channels[0]
	assert.Equal(t, "chat-room-(?P<id>\\d+)", cfg.Regex.String())
	assert.Equal(t, MakeDurationPtr("30m"), cfg.Expire)

	assert.Equal(t, 0, len(cfg.Webhooks))
	assert.Nil(t, cfg.Jwt)
}

func TestChannelWebhookDefaultConfig(t *testing.T) {
	configYaml := strings.ReplaceAll(`
channels:
-
	regex: 'chat-room-(?P<id>\d+)'
	webhooks:
		-
			url: 'http://localhost:3001/you-got-message/room/{{.id}}'
`, "\t", "  ")
	config, err := ParseConfig(Overrides{}, configYaml)
	if err != nil {
		t.Error(err)
		return
	}

	cfg := config.Channels[0]
	assert.Equal(t, "chat-room-(?P<id>\\d+)", cfg.Regex.String())

	webhook := cfg.Webhooks[0]
	assert.Equal(t, "http://localhost:3001/you-got-message/room/{{.id}}", webhook.URL.String())
	assert.Equal(t, MakeDurationPtr("30s"), webhook.Timeout)
	assert.Equal(t, MakeIntPtr(3), webhook.Retry.Count)
	assert.Equal(t, MakeDurationPtr("3s"), webhook.Retry.Interval)
	assert.Equal(t, 1.5, *webhook.Retry.IntervalMultiplier)
	assert.Equal(t, MakeDurationPtr("1.5s"), webhook.Retry.IntervalJitter)
	assert.Equal(t, 0, len(webhook.Headers))
}

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
			- |-
				-----BEGIN PUBLIC KEY-----
				MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAlgY7fpgGEKqGaoUc1O9K
				CdytNmBa7P1DWfA8QWFE042yn/dBLW8M+uWqsvD/pDWaSDNfEgY6J8nyKZ7DMps6
				E1TJNBkZ7/4TDVpmsIE8vqK/bhTz5SYTnLyMd2Wh7Yy+uUOk6XTR2Ade9ysHPD5U
				mmFBzQX2r+S25lpRUHXmSGl7cYiTbWmI2JVTId3agHR1jqZ1EeWDorEZ3HF7hExl
				pKXa0vZaMoK2mvzHOhaNPn57BNqcXfzLVYjny1br7qJOHgMBW+AwCbb7yE+aRsur
				WQEc6XyhbFG443Sb6tHvbiROg2nTXu1Pq0ZaB90mytpm0Md+p0QI0mqizhbOD3d3
				Lf10Zj86nlvT4dKbWwZHfrh9oiR9tLGgCyUtVQYhgv7BehdLnpJmxxaohteLJHon
				PfzIKqOY24OmteqAML7+G8gbrRIXMS8aTvPJvJ3XT51QD+61CMwExMWXz1CTXlc3
				tSZ0nx8hquPI9C/B9AIlnk0lgKNmq+A2aU98OnSlTPqsdZo3xr4PPMthiNr/dfEq
				HsijJ3dq9pwaO9t0xKti+Hd9ic/IqUH2OyT0Nw36f/MvDBAILF8SVimSKnEaQI04
				5AME2BK5WZiwL47SqZIWTNUglhyPEZCZ2tFJYHZHFSW6AbnDWAxYKBuDE7MB+t/u
				Y4XfEnmCs8dK48LUuB+IgF8CAwEAAQ==
				-----END PUBLIC KEY-----
`, "\t", "  ")
	config, err := ParseConfig(Overrides{}, configYaml)
	if err != nil {
		t.Error(err)
		return
	}

	cfg := config.Channels[0]
	assert.Equal(t, "chat-room-(?P<id>\\d+)", cfg.Regex.String())

	jwt := cfg.Jwt
	assert.Equal(t, 0, len(jwt.Claims))
}

func TestChannelFullConfig(t *testing.T) {
	configYaml := strings.ReplaceAll(`
channels:
-
	regex: 'chat-room-(?P<id>\d+)'
	# Must be larger than final retry attempt time
	expire: 15m
	webhooks:
		-
			url: 'http://localhost:3001/you-got-message/room/{{.id}}'
			timeout: 61s
			retry:
				count: 4
				interval: 3.5s
				intervalMultiplier: 3.1
				intervalJitter: 2s500ms
			headers:
				User-Agent: my DSPS server
				X-Chat-Room-ID: '{{.id}}'
	jwt:
		alg: RS256
		iss:
			- https://issuer.example.com/issuer-url
		keys:
			- |-
				-----BEGIN PUBLIC KEY-----
				MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAlgY7fpgGEKqGaoUc1O9K
				CdytNmBa7P1DWfA8QWFE042yn/dBLW8M+uWqsvD/pDWaSDNfEgY6J8nyKZ7DMps6
				E1TJNBkZ7/4TDVpmsIE8vqK/bhTz5SYTnLyMd2Wh7Yy+uUOk6XTR2Ade9ysHPD5U
				mmFBzQX2r+S25lpRUHXmSGl7cYiTbWmI2JVTId3agHR1jqZ1EeWDorEZ3HF7hExl
				pKXa0vZaMoK2mvzHOhaNPn57BNqcXfzLVYjny1br7qJOHgMBW+AwCbb7yE+aRsur
				WQEc6XyhbFG443Sb6tHvbiROg2nTXu1Pq0ZaB90mytpm0Md+p0QI0mqizhbOD3d3
				Lf10Zj86nlvT4dKbWwZHfrh9oiR9tLGgCyUtVQYhgv7BehdLnpJmxxaohteLJHon
				PfzIKqOY24OmteqAML7+G8gbrRIXMS8aTvPJvJ3XT51QD+61CMwExMWXz1CTXlc3
				tSZ0nx8hquPI9C/B9AIlnk0lgKNmq+A2aU98OnSlTPqsdZo3xr4PPMthiNr/dfEq
				HsijJ3dq9pwaO9t0xKti+Hd9ic/IqUH2OyT0Nw36f/MvDBAILF8SVimSKnEaQI04
				5AME2BK5WZiwL47SqZIWTNUglhyPEZCZ2tFJYHZHFSW6AbnDWAxYKBuDE7MB+t/u
				Y4XfEnmCs8dK48LUuB+IgF8CAwEAAQ==
				-----END PUBLIC KEY-----
		claims:
			chatroom: '{{.id}}'
`, "\t", "  ")
	config, err := ParseConfig(Overrides{}, configYaml)
	if err != nil {
		t.Error(err)
		return
	}

	cfg := config.Channels[0]
	assert.Equal(t, "chat-room-(?P<id>\\d+)", cfg.Regex.String())
	assert.Equal(t, MakeDurationPtr("15m"), cfg.Expire)

	webhook := cfg.Webhooks[0]
	assert.Equal(t, "http://localhost:3001/you-got-message/room/{{.id}}", webhook.URL.String())
	assert.Equal(t, MakeDurationPtr("61s"), webhook.Timeout)
	assert.Equal(t, MakeIntPtr(4), webhook.Retry.Count)
	assert.Equal(t, MakeDurationPtr("3.5s"), webhook.Retry.Interval)
	assert.Equal(t, 3.1, *webhook.Retry.IntervalMultiplier)
	assert.Equal(t, MakeDurationPtr("2.5s"), webhook.Retry.IntervalJitter)
	assert.Equal(t, "my DSPS server", webhook.Headers["User-Agent"].String())
	assert.Equal(t, "{{.id}}", webhook.Headers["X-Chat-Room-ID"].String())

	jwt := cfg.Jwt
	assert.Equal(t, "RS256", jwt.Alg)
	assert.Equal(t, []string{"https://issuer.example.com/issuer-url"}, jwt.Iss)
	assert.Equal(t, `-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAlgY7fpgGEKqGaoUc1O9K
CdytNmBa7P1DWfA8QWFE042yn/dBLW8M+uWqsvD/pDWaSDNfEgY6J8nyKZ7DMps6
E1TJNBkZ7/4TDVpmsIE8vqK/bhTz5SYTnLyMd2Wh7Yy+uUOk6XTR2Ade9ysHPD5U
mmFBzQX2r+S25lpRUHXmSGl7cYiTbWmI2JVTId3agHR1jqZ1EeWDorEZ3HF7hExl
pKXa0vZaMoK2mvzHOhaNPn57BNqcXfzLVYjny1br7qJOHgMBW+AwCbb7yE+aRsur
WQEc6XyhbFG443Sb6tHvbiROg2nTXu1Pq0ZaB90mytpm0Md+p0QI0mqizhbOD3d3
Lf10Zj86nlvT4dKbWwZHfrh9oiR9tLGgCyUtVQYhgv7BehdLnpJmxxaohteLJHon
PfzIKqOY24OmteqAML7+G8gbrRIXMS8aTvPJvJ3XT51QD+61CMwExMWXz1CTXlc3
tSZ0nx8hquPI9C/B9AIlnk0lgKNmq+A2aU98OnSlTPqsdZo3xr4PPMthiNr/dfEq
HsijJ3dq9pwaO9t0xKti+Hd9ic/IqUH2OyT0Nw36f/MvDBAILF8SVimSKnEaQI04
5AME2BK5WZiwL47SqZIWTNUglhyPEZCZ2tFJYHZHFSW6AbnDWAxYKBuDE7MB+t/u
Y4XfEnmCs8dK48LUuB+IgF8CAwEAAQ==
-----END PUBLIC KEY-----`, jwt.Keys[0])
	assert.Equal(t, "{{.id}}", jwt.Claims["chatroom"].String())
}
