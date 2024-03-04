package telemetry

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc/credentials"
	"log"
)

func StartTracer(ctx context.Context, app, version string) func(context.Context) error {
	secureOption := otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if len(insecure) > 0 {
		secureOption = otlptracegrpc.WithInsecure()
	}

	exporter, err := otlptrace.New(ctx,
		otlptracegrpc.NewClient(secureOption, otlptracegrpc.WithEndpoint(collectorURL)),
	)
	if err != nil {
		log.Printf("unable to create otlp trace exporter: %s\n", err)
		return func(ctx context.Context) error { return nil }
	}

	resources := resource.NewWithAttributes(semconv.SchemaURL,
		semconv.ServiceName(app),
		semconv.ServiceVersion(version),
		semconv.TelemetrySDKLanguageGo)

	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(resources),
		),
	)
	return exporter.Shutdown
}
