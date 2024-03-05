# go-telemetry

A standard library for instrumenting Go apps with OpenTelemetry

This library is a batteries-included setup to make it easy for Go apps to get up and running with 3 lines of code.
If you want to highly customize your OpenTelemetry setup, this library is a helpful reference for setup.

## How to use

```go
func main() {
    shutdownTelemetry := telemetry.Start(ctx)
    defer shutdownTelemetry()
    telemetry.WatchRuntime()
	
	// Start server, perform job, etc.
}
```

## OpenTelemetry Configuration

OpenTelemetry provides a set of standard environment variables that the official Golang libraries use to configure.
(See [Environment Variable Specification](https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/))

### Tracer Exporters

This library uses automatic detection of tracer exporters using `OTEL_TRACES_EXPORTER` environment variable.
By default, this supports:
- "otlp": [OTLP](https://opentelemetry.io/docs/specs/otel/protocol/otlp/)
- "console": [Standard Output](https://opentelemetry.io/docs/specs/otel/trace/sdk_exporters/stdout/)
- "none": No exporter

#### OTLP Tracer (production)

```dotenv
OTEL_TRACES_EXPORTER: otlp
```

#### OTLP Tracer (local docker)

Configure OTLP exporter to emit traces to an OTLP server.
This example shows how to configure an OTLP endpoint from within a docker container to the host machine.

```dotenv
OTEL_TRACES_EXPORTER: otlp
OTEL_EXPORTER_OTLP_ENDPOINT: host.docker.internal:4317
OTEL_EXPORTER_OTLP_INSECURE: true
```

#### Console Tracer

Configure console exporter to emit traces to stdout.

```dotenv
OTEL_TRACES_EXPORTER: console
```

### Tracer Sampling

Configure trace sampling with `OTEL_TRACES_SAMPLER` environment variable.
Here are the choices.
- "always_on": AlwaysOnSampler
- "always_off": AlwaysOffSampler
- "traceidratio": TraceIdRatioBased
- "parentbased_always_on": ParentBased(root=AlwaysOnSampler)
- "parentbased_always_off": ParentBased(root=AlwaysOffSampler)
- "parentbased_traceidratio": ParentBased(root=TraceIdRatioBased)
- "parentbased_jaeger_remote": ParentBased(root=JaegerRemoteSampler)
- "jaeger_remote": JaegerRemoteSampler
- "xray": AWS X-Ray Centralized Sampling (third party)

### Metric Readers

This library uses automatic detection of metric readers using `OTEL_METRICS_EXPORTER` environment variable.
By default, this supports:
- "otlp": [OTLP](https://opentelemetry.io/docs/specs/otel/protocol/otlp/)
- "console": [Standard Output](https://opentelemetry.io/docs/specs/otel/metrics/sdk_exporters/stdout/)
- "none": No exporter
- "prometheus": [Prometheus](https://github.com/prometheus/docs/blob/master/content/docs/instrumenting/exposition_formats.md)
