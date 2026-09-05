package telemetry

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

func StartLogger(ctx context.Context, resources *resource.Resource, cfg Config) func(context.Context) error {
	if !cfg.EnableLogger {
		return func(ctx context.Context) error { return nil }
	}

	// OTEL_LOGS_EXPORTER selects the exporter: "otlp" (default), "console", or "none" -- matching
	// how StartTracer and StartMetrics are configured. Don't add a second, hardcoded stdout
	// exporter here: it pretty-prints every record to stdout on top of the OTLP export, which in
	// production doubles log volume and cannot be turned off.
	exporter, err := autoexport.NewLogExporter(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to create otlp log exporter: %s\n", err)
		return func(ctx context.Context) error { return nil }
	}

	provider := log.NewLoggerProvider(
		log.WithResource(resources),
		log.WithProcessor(log.NewBatchProcessor(exporter)),
	)
	global.SetLoggerProvider(provider)

	return provider.Shutdown
}
