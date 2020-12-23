package config

import (
	"fmt"

	"github.com/saiya/dsps/server/domain"
)

// TelemetryConfig is to setup tracing/metrics.
type TelemetryConfig struct {
	OT *OpenTelemetryConfig `json:"ot"`
}

// OpenTelemetryConfig is to setup OpenTelemetry
type OpenTelemetryConfig struct {
	Tracing *OpenTelemetryTracingConfig `json:"tracing"`

	Exporters OpenTelemetryExportersConfig `json:"exporters"`
}

// OpenTelemetryTracingConfig is to setup OpenTelemetry tracing
type OpenTelemetryTracingConfig struct {
	Enable   bool    `json:"enable"`
	Sampling float64 `json:"sampling"`

	Batch OpenTelemetryTracingBatchConfig `json:"batch"`

	Attributes map[string]interface{} `json:"attributes"`
}

// OpenTelemetryTracingBatchConfig is to configure batch transmission
type OpenTelemetryTracingBatchConfig struct {
	MaxQueueSize *int             `json:"maxQueueSize"`
	Timeout      *domain.Duration `json:"timeout"`
	BatchSize    *int             `json:"batchSize"`
}

// OpenTelemetryExportersConfig configure exporters
type OpenTelemetryExportersConfig struct {
	Stdout OpenTelemetryExporterStdoutConfig `json:"stdout"`
	GCP    OpenTelemetryExporterGCPConfig    `json:"gcp"`
}

// OpenTelemetryExporterStdoutConfig configure stdout exporter
type OpenTelemetryExporterStdoutConfig struct {
	Enable    bool      `json:"enable"`
	Quantiles []float64 `json:"quantiles"`
}

// OpenTelemetryExporterGCPConfig configure GCP expoter
type OpenTelemetryExporterGCPConfig struct {
	EnableTrace bool   `json:"enableTrace"`
	ProjectID   string `json:"projectID"`
}

func tracingConfigDefault() *TelemetryConfig {
	return &TelemetryConfig{
		OT: &OpenTelemetryConfig{
			Tracing: &OpenTelemetryTracingConfig{Enable: false},
		},
	}
}

// PostprocessTelemetryConfig fixes/validates config
func PostprocessTelemetryConfig(config *TelemetryConfig) error {
	if err := postprocessOTTracingConfig(config.OT.Tracing); err != nil {
		return fmt.Errorf(`OT tracing configuration error: %w`, err)
	}

	if len(config.OT.Exporters.Stdout.Quantiles) == 0 {
		config.OT.Exporters.Stdout.Quantiles = []float64{0.5, 0.9, 0.99}
	}

	return nil
}

func postprocessOTTracingConfig(config *OpenTelemetryTracingConfig) error {
	if config.Batch.MaxQueueSize == nil {
		config.Batch.MaxQueueSize = makeIntPtr(2048)
	}
	if config.Batch.Timeout == nil {
		config.Batch.Timeout = makeDurationPtr("5s")
	}
	if config.Batch.BatchSize == nil {
		config.Batch.BatchSize = makeIntPtr(512)
	}

	if config.Sampling < 0 || 1 < config.Sampling {
		return fmt.Errorf(`sampling ration must be within [0.0, 1.0]`)
	}

	return nil
}
