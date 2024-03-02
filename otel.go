package telemetry

import (
	"os"
)

var (
	collectorURL = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	insecure     = os.Getenv("OTEL_INSECURE_MODE")
)
