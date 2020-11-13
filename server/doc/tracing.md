# Distribution Tracing x DSPS

DSPS server supports [OpenTelemetry](https://opentelemetry.io/).

## Exporters

### GCP Cloud Trace

If you specify `GOOGLE_CLOUD_PROJECT` environment variable, DSPS server activates [Cloud Trace exporter](https://cloud.google.com/trace/docs/setup/go-ot).

So that you can collect traces into GCP Cloud Trace.
