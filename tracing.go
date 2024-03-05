package telemetry

import (
	"context"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"log"
)

func StartTracer(ctx context.Context, resources *resource.Resource) func(context.Context) error {
	exporter, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		log.Printf("unable to create otlp trace exporter: %s\n", err)
		return func(ctx context.Context) error { return nil }
	}
	provider := trace.NewTracerProvider(
		trace.WithResource(resources),
		trace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(provider)
	return provider.Shutdown
}
