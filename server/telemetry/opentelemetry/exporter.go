package opentelemetry

import (
	gcpexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/otel/exporters/stdout"
	exporttrace "go.opentelemetry.io/otel/sdk/export/trace"
	"golang.org/x/xerrors"

	"github.com/saiya/dsps/server/config"
)

func newExporters(config *config.OpenTelemetryConfig) ([]exporttrace.SpanExporter, error) {
	exporters := make([]exporttrace.SpanExporter, 0, 1)
	if config.Exporters.Stdout.Enable {
		exporter, err := stdout.NewExporter(
			stdout.WithQuantiles(config.Exporters.Stdout.Quantiles),
		)
		if err != nil {
			return nil, xerrors.Errorf("failed to initialize OpenTelemetry stdout trace exporter: %w", err)
		}
		exporters = append(exporters, exporter)
	}
	if config.Exporters.GCP.EnableTrace {
		exporter, err := gcpexporter.NewExporter(gcpexporter.WithProjectID(config.Exporters.GCP.ProjectID))
		if err != nil {
			return nil, xerrors.Errorf("failed to initialize OpenTelemetry GCP Cloud Trace exporter: %w", err)
		}
		exporters = append(exporters, exporter)
	}
	return exporters, nil
}
