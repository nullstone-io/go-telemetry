# go-telemetry
A standard library for instrumenting golang apps with OpenTelemetry

## How to use

```go
func main() {
    cleanupTelemetry := telemetry.Start(ctx, AppName, Version)
    defer cleanupTelemetry()
    telemetry.WatchRuntime()
	
	// Start server, perform job, etc.
}
```
