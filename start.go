package telemetry

import (
	"context"
	"errors"
	"go.opentelemetry.io/otel/sdk/resource"
	"log"
)

func Start(ctx context.Context) func() {
	resources, err := DetectResources(ctx)
	if err != nil {
		log.Fatalf("Error initializing OTEL resources: %s\n", err)
	}

	shutdownTracer := StartTracer(ctx, resources)
	shutdownMetrics := StartMetrics(ctx, resources)
	return func() {
		shutdownTracer(ctx)
		shutdownMetrics(ctx)
	}
}

func DetectResources(ctx context.Context) (*resource.Resource, error) {
	res, err := resource.New(
		ctx,
		resource.WithDetectors(NullstoneResourceDetector{}), // Nullstone resource detection
		resource.WithFromEnv(),                              // Discover and provide attributes from OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME environment variables.
		resource.WithTelemetrySDK(),                         // Discover and provide information about the OpenTelemetry SDK used.
		resource.WithProcess(),                              // Discover and provide process information.
		resource.WithOS(),                                   // Discover and provide OS information.
		resource.WithContainer(),                            // Discover and provide container information.
		resource.WithHost(),                                 // Discover and provide host information.
		//resource.WithAttributes(attribute.String("foo", "bar")), // Add custom resource attributes.
	)
	if errors.Is(err, resource.ErrPartialResource) || errors.Is(err, resource.ErrSchemaURLConflict) {
		// Log non-fatal issues
		log.Printf("Error initializing detecting OTEL resources: %s\n", err)
		return res, nil
	} else if err != nil {
		// This is a fatal issue and should fail the main program
		return nil, err
	}
	return res, nil
}
