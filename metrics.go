package telemetry

import (
	"context"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"log"
)

func StartMetrics(ctx context.Context, resources *resource.Resource) func(ctx context.Context) error {
	reader, err := autoexport.NewMetricReader(ctx)
	if err != nil {
		log.Printf("error creating otlp metric exporter: %s\n", err)
		return func(ctx context.Context) error { return nil }
	}
	provider := metric.NewMeterProvider(
		metric.WithResource(resources),
		metric.WithReader(reader),
	)
	otel.SetMeterProvider(provider)
	return provider.Shutdown
}
