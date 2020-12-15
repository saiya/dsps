package config_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/config"
	"github.com/saiya/dsps/server/domain"
)

func TestHttpServerDefaultValues(t *testing.T) {
	config, err := ParseConfig(context.Background(), Overrides{}, ``)
	if err != nil {
		t.Error(err)
		return
	}

	cfg := *config.HTTPServer
	assert.Equal(t, "", cfg.RealIPHeader)
	assert.Equal(t, len(domain.PrivateCIDRs), len(cfg.TrustedProxyRanges))
	assert.Equal(t, domain.PrivateCIDRs[0].String(), cfg.TrustedProxyRanges[0].String())
	assert.Equal(t, `deny`, cfg.DefaultHeaders["X-Frame-Options"])
}

func TestHttpServerNonDefaultValues(t *testing.T) {
	configYaml := strings.ReplaceAll(`
http:
	realIpHeader: X-Forwarded-For
	trustedProxyRanges:
		- 1.2.3.4/16
	defaultHeaders:
		X-Frame-Options:
`, "\t", "  ")
	config, err := ParseConfig(context.Background(), Overrides{}, configYaml)
	if err != nil {
		t.Error(err)
		return
	}

	cfg := *config.HTTPServer
	assert.Equal(t, "X-Forwarded-For", cfg.RealIPHeader)
	assert.Equal(t, 1, len(cfg.TrustedProxyRanges))
	assert.Equal(t, "1.2.3.4/16", cfg.TrustedProxyRanges[0].String())
	assert.Equal(t, ``, cfg.DefaultHeaders["X-Frame-Options"])
}

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
