package telemetry

import (
	"context"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc/credentials"
	"log"
	"time"
)

const (
	meterInterval = 1 * time.Second
)

func StartMetrics(ctx context.Context, app, version string) func(ctx context.Context) error {
	secureOption := otlpmetricgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if len(insecure) > 0 {
		secureOption = otlpmetricgrpc.WithInsecure()
	}

	exporter, err := otlpmetricgrpc.New(ctx, secureOption, otlpmetricgrpc.WithEndpoint(collectorURL))
	if err != nil {
		log.Printf("error creating otlp metric exporter: %s\n", err)
		return func(ctx context.Context) error { return nil }
	}

	resources := resource.NewWithAttributes(semconv.SchemaURL,
		semconv.ServiceName(app),
		semconv.ServiceVersion(version),
		semconv.TelemetrySDKLanguageGo)

	read := metric.NewPeriodicReader(exporter, metric.WithInterval(meterInterval))
	provider := metric.NewMeterProvider(metric.WithResource(resources), metric.WithReader(read))
	return provider.Shutdown
}
