package config_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/config"
)

func TestSentryDefaults(t *testing.T) {
	hostname, err := os.Hostname()
	assert.NoError(t, err)
	assert.NotEmpty(t, hostname)

	withSentryDsn(t, "", func() {
		config, err := ParseConfig(context.Background(), Overrides{
			BuildVersion: "v9.8.7",
			BuildDist:    "someOS-arch",
		}, "sentry: {}")
		assert.NoError(t, err)

		sentry := config.Sentry
		assert.Equal(t, "", sentry.DSN)
		assert.Equal(t, hostname, sentry.ServerName)
		assert.Equal(t, "v9.8.7", sentry.Release)
		assert.Equal(t, "someOS-arch", sentry.Distribution)
		assert.Equal(t, 1.0, *sentry.SampleRate)
		assert.Equal(t, 15*time.Second, sentry.FlushTimeout.Duration)
	})
}

func TestSentryConfigDump(t *testing.T) {
	dsn := "DSN89E0B2588C904EB1A80F994545F5A1B7DSN"
	withSentryDsn(t, dsn, func() {
		config, err := ParseConfig(context.Background(), Overrides{}, "")
		assert.NoError(t, err)

		dumpB := strings.Builder{}
		assert.NoError(t, config.DumpConfig(&dumpB))
		dump := dumpB.String()
		assert.NotEmpty(t, dump)
		assert.NotContains(t, dump, dsn) // DSN (contains credential) should not be dumped
	})
}

func TestSentryConfigError(t *testing.T) {
	_, err := ParseConfig(context.Background(), Overrides{}, `sentry: { sampleRate: -1.0 }`)
	assert.Regexp(t, `Sentry configration problem: sample ratio must be within \[0.0, 1.0\]`, err.Error())

	_, err = ParseConfig(context.Background(), Overrides{}, `sentry: { flushTimeout: 0s }`)
	assert.Regexp(t, `flushTimeout must be larger than zero`, err.Error())
}

func withSentryDsn(t *testing.T, dsn string, f func()) {
	old := os.Getenv("SENTRY_DSN")
	defer func() { assert.NoError(t, os.Setenv("SENTRY_DSN", old)) }()

	assert.NoError(t, os.Setenv("SENTRY_DSN", dsn))
	f()
}
