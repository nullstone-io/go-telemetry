package telemetry

import (
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"time"
)

func WatchRuntime() error {
	return runtime.Start(runtime.WithMinimumReadMemStatsInterval(5 * time.Second))
}
