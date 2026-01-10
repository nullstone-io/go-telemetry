package telemetry

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

func StartLogger(ctx context.Context, resources *resource.Resource, cfg Config) func(context.Context) error {
	if !cfg.EnableLogger {
		return func(ctx context.Context) error { return nil }
	}

	exporter, err := autoexport.NewLogExporter(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to create otlp log exporter: %s\n", err)
		return func(ctx context.Context) error { return nil }
	}
	stdoutExporter, err := stdoutlog.New(
		stdoutlog.WithPrettyPrint(),
		stdoutlog.WithoutTimestamps(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to create stdout log exporter: %s\n", err)
		return func(ctx context.Context) error { return nil }
	}

	provider := log.NewLoggerProvider(
		log.WithResource(resources),
		log.WithProcessor(log.NewSimpleProcessor(stdoutExporter)),
		log.WithProcessor(log.NewBatchProcessor(exporter)),
	)
	global.SetLoggerProvider(provider)

	return provider.Shutdown
}
