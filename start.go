package telemetry

import (
	"context"
	"os"
)

var (
	AppName = os.Getenv("NULLSTONE_APP")
	Version = os.Getenv("NULLSTONE_VERSION")
)

func Start(ctx context.Context, app, version string) func() {
	if app == "" {
		app = AppName
	}
	if version == "" {
		version = Version
	}
	shutdownMetrics := StartMetrics(ctx, app, version)
	shutdownTracer := StartTracer(ctx, app, version)
	return func() {
		shutdownTracer(ctx)
		shutdownMetrics(ctx)
	}
}
