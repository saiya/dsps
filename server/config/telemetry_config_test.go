package config_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/config"
)

func TestTracingEmptyConfig(t *testing.T) {
	config, err := ParseConfig(context.Background(), Overrides{}, ``)
	if err != nil {
		t.Error(err)
		return
	}

	assert.NotNil(t, config.Telemetry)
	ott := config.Telemetry.OT.Tracing
	assert.NotNil(t, ott)
	assert.False(t, ott.Enable)
	assert.False(t, config.Telemetry.OT.Exporters.Stdout.Enable)
}

func TestTracingFullConfig(t *testing.T) {
	configYaml := strings.ReplaceAll(`
telemetry:
	ot:
		tracing:
			enable: true
			sampling: 0.14

			batch:
				maxQueueSize: 12
				timeout: 34s
				batchSize: 56
			
			attributes:
				str: foo bar
				bool: true
				int: 1234
				float: 3.14
		
		exporters:
			stdout:
				enable: true
`, "\t", "  ")
	config, err := ParseConfig(context.Background(), Overrides{}, configYaml)
	if err != nil {
		t.Error(err)
		return
	}

	ott := config.Telemetry.OT.Tracing
	assert.True(t, ott.Enable)
	assert.Equal(t, 0.14, ott.Sampling)

	assert.Equal(t, 12, *ott.Batch.MaxQueueSize)
	assert.Equal(t, 34*time.Second, ott.Batch.Timeout.Duration)
	assert.Equal(t, 56, *ott.Batch.BatchSize)

	assert.Equal(t, "foo bar", ott.Attributes["str"])
	assert.Equal(t, true, ott.Attributes["bool"])
	assert.Equal(t, uint64(1234), ott.Attributes["int"]) // Should be parsed as integer
	assert.Equal(t, 3.14, ott.Attributes["float"])

	assert.True(t, config.Telemetry.OT.Exporters.Stdout.Enable)
}

func TestTracingConfigError(t *testing.T) {
	_, err := ParseConfig(context.Background(), Overrides{}, `telemetry: { ot: { tracing: { enable: true, sampling: -2.0 } } }`)
	assert.Regexp(t, `sampling ration must be within \[0.0, 1.0\]`, err.Error())
}
