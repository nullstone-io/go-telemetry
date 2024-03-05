package telemetry

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"os"
)

var (
	StackName = os.Getenv("NULLSTONE_STACK")
	EnvName   = os.Getenv("NULLSTONE_ENV")
	AppName   = os.Getenv("NULLSTONE_APP")
	Version   = os.Getenv("NULLSTONE_VERSION")
)

var _ resource.Detector = NullstoneResourceDetector{}

type NullstoneResourceDetector struct{}

func (n NullstoneResourceDetector) Detect(ctx context.Context) (*resource.Resource, error) {
	attrs := make([]attribute.KeyValue, 0)
	if StackName != "" {
		attrs = append(attrs, attribute.String("nullstone.stack", StackName))
	}
	if EnvName != "" {
		attrs = append(attrs, semconv.DeploymentEnvironment(EnvName))
	}
	if AppName != "" {
		attrs = append(attrs, semconv.ServiceName(AppName))
	}
	if Version != "" {
		attrs = append(attrs, semconv.ServiceVersion(Version))
	}

	return resource.NewSchemaless(attrs...), nil
}
