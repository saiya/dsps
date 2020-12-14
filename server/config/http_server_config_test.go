package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/config"
)

func TestHttpServerConfigOverride(t *testing.T) {
	cfg := HTTPServerConfig{}
	assert.NoError(t, PostprocessHTTPServerConfig(&cfg, Overrides{Port: 9876}))
	assert.Equal(t, 9876, cfg.Port)
	assert.Equal(t, ":9876", cfg.Listen)

	cfg = HTTPServerConfig{}
	assert.NoError(t, PostprocessHTTPServerConfig(&cfg, Overrides{Listen: "127.0.0.1:8765"}))
	assert.Equal(t, "127.0.0.1:8765", cfg.Listen)
}

func TestHttpServerConfigPathPrefix(t *testing.T) {
	for _, testcase := range []struct{ source, expected string }{
		{source: "", expected: ""},
		{source: "/", expected: ""},
		{source: "foo/bar", expected: "foo/bar"},
		{source: "foo/bar/", expected: "foo/bar"},
		{source: "/foo/bar", expected: "foo/bar"},
		{source: "/foo/bar/", expected: "foo/bar"},
	} {
		cfg := HTTPServerConfig{PathPrefix: testcase.source}
		assert.NoError(t, PostprocessHTTPServerConfig(&cfg, Overrides{Port: 9876}))
		assert.Equal(t, testcase.expected, cfg.PathPrefix)
	}
}

func TestDefaultHeaders(t *testing.T) {
	cfg := HTTPServerConfig{}
	assert.NoError(t, PostprocessHTTPServerConfig(&cfg, Overrides{}))
	assert.Equal(t, `nosniff`, cfg.DefaultHeaders["X-Content-Type-Options"]) // default

	cfg = HTTPServerConfig{
		// Values from config file
		DefaultHeaders: map[string]string{
			"X-Content-Type-Options":    ``, // disable this header
			"Strict-Transport-Security": `max-age=31536000 ; includeSubDomains`,
		},
	}
	assert.NoError(t, PostprocessHTTPServerConfig(&cfg, Overrides{}))
	assert.Equal(t, ``, cfg.DefaultHeaders["X-Content-Type-Options"])
	assert.Equal(t, `max-age=31536000 ; includeSubDomains`, cfg.DefaultHeaders["Strict-Transport-Security"])
	assert.Equal(t, `no-cache`, cfg.DefaultHeaders["Pragma"]) // default
}
