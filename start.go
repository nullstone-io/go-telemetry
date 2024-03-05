package telemetry

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"os"
	"slices"
)

var (
	AppName = os.Getenv("NULLSTONE_APP")
	EnvName = os.Getenv("NULLSTONE_ENV")
	Version = os.Getenv("NULLSTONE_VERSION")
)

func Start(ctx context.Context, attrs ...attribute.KeyValue) func() {
	effective := slices.Clone(attrs)
	if !attrSliceContainsKey(effective, semconv.ServiceNameKey) {
		effective = append(effective, semconv.ServiceName(AppName))
	}
	if !attrSliceContainsKey(effective, semconv.DeploymentEnvironmentKey) {
		effective = append(effective, semconv.DeploymentEnvironment(EnvName))
	}
	if !attrSliceContainsKey(effective, semconv.ServiceVersionKey) {
		effective = append(effective, semconv.ServiceVersion(Version))
	}

	shutdownMetrics := StartMetrics(ctx, effective...)
	shutdownTracer := StartTracer(ctx, effective...)
	return func() {
		shutdownTracer(ctx)
		shutdownMetrics(ctx)
	}
}

func attrSliceContainsKey(s []attribute.KeyValue, key attribute.Key) bool {
	return slices.ContainsFunc(s, func(value attribute.KeyValue) bool {
		return value.Key == key
	})
}
