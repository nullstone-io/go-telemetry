package telemetry

import "context"

func Init(ctx context.Context, app, version string) func() {
	cleanupMetrics := InitMetrics(ctx, app, version)
	cleanupTracer := InitTracer(ctx, app, version)
	return func() {
		cleanupTracer(ctx)
		cleanupMetrics(ctx)
	}
}
