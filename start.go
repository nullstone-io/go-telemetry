package telemetry

import "context"

func Start(ctx context.Context, app, version string) func() {
	shutdownMetrics := StartMetrics(ctx, app, version)
	shutdownTracer := StartTracer(ctx, app, version)
	return func() {
		shutdownTracer(ctx)
		shutdownMetrics(ctx)
	}
}
