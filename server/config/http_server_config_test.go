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
		{source: "", expected: "/"},
		{source: "/", expected: "/"},
		{source: "foo/bar", expected: "/foo/bar"},
		{source: "foo/bar/", expected: "/foo/bar"},
		{source: "/foo/bar", expected: "/foo/bar"},
		{source: "/foo/bar/", expected: "/foo/bar"},
	} {
		cfg := HTTPServerConfig{PathPrefix: testcase.source}
		assert.NoError(t, PostprocessHTTPServerConfig(&cfg, Overrides{Port: 9876}))
		assert.Equal(t, testcase.expected, cfg.PathPrefix)
	}
}
