package workflows

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
)

// Span attribute keys recorded on a failed execution's span
const (
	AttrClass     = "workflow.failure.class"
	AttrCategory  = "workflow.failure.category"
	AttrCode      = "workflow.failure.code"
	AttrErrorType = "error.type"
)

// Attributes returns the span attributes for this classification; errType is the Go type of the underlying error
func (i Info) Attributes(errType string) []attribute.KeyValue {
	attrs := []attribute.KeyValue{attribute.String(AttrClass, string(i.Class))}
	if i.Category != "" {
		attrs = append(attrs, attribute.String(AttrCategory, i.Category))
	}
	if i.Code != "" {
		attrs = append(attrs, attribute.String(AttrCode, i.Code))
	}
	if errType != "" {
		attrs = append(attrs, attribute.String(AttrErrorType, errType))
	}
	return attrs
}

// Annotate records the classification of err on span
func Annotate(span trace.Span, info Info, err error) {
	span.SetAttributes(info.Attributes(ErrorType(err))...)
}

// Metric names
const (
	MetricWorkflowCompletions = "workflow.completions"
	MetricWorkflowFailures    = "workflow.failures"
	MetricActivityFailures    = "workflow.activity.failures"
)

// Metric attribute keys. Deliberately low-cardinality: org/stack/env stay on spans only.
const (
	MetricAttrStatus       = "status"
	MetricAttrClass        = "class"
	MetricAttrCategory     = "category"
	MetricAttrWorkflowType = "workflow_type"
	MetricAttrActivityType = "activity_type"
	// MetricAttrRoot is true for a workflow with no parent, i.e. one count per customer-visible intent
	MetricAttrRoot = "root"
)

// Completion statuses for MetricWorkflowCompletions
const (
	CompletionSucceeded = "succeeded"
	CompletionFailed    = "failed"
	CompletionCancelled = "cancelled"
	CompletionTimeout   = "timeout"
)

const instrumentationName = "github.com/nullstone-io/go-telemetry/workflows"

var (
	workflowCompletions metric.Int64Counter = noop.Int64Counter{}
	workflowFailures    metric.Int64Counter = noop.Int64Counter{}
	activityFailures    metric.Int64Counter = noop.Int64Counter{}
)

func init() {
	// The global MeterProvider delegates to the provider telemetry.Start installs later, so instruments can be created at init
	meter := otel.Meter(instrumentationName)
	workflowCompletions = counter(meter, MetricWorkflowCompletions, "Workflow executions that finished, by final status", "{workflow}")
	workflowFailures = counter(meter, MetricWorkflowFailures, "Workflow executions that failed, by failure class and category", "{workflow}")
	activityFailures = counter(meter, MetricActivityFailures, "Activity executions that failed, by failure class and category", "{activity}")
}

func counter(meter metric.Meter, name, description, unit string) metric.Int64Counter {
	c, err := meter.Int64Counter(name, metric.WithDescription(description), metric.WithUnit(unit))
	if err != nil || c == nil {
		return noop.Int64Counter{}
	}
	return c
}

// CompletionStatus maps a classification to the status recorded on MetricWorkflowCompletions
func CompletionStatus(info Info) string {
	switch info.Class {
	case ClassCancelled:
		return CompletionCancelled
	case ClassTimeout:
		return CompletionTimeout
	}
	return CompletionFailed
}

// RecordWorkflowCompletion counts a finished workflow by status and, when it failed (user or internal),
// by class and category. info is ignored when err is nil.
func RecordWorkflowCompletion(ctx context.Context, workflowType string, root bool, info Info, err error) {
	status := CompletionSucceeded
	if err != nil {
		status = CompletionStatus(info)
	}
	workflowCompletions.Add(ctx, 1, metric.WithAttributes(
		attribute.String(MetricAttrStatus, status),
		attribute.String(MetricAttrWorkflowType, workflowType),
	))
	if err != nil && info.Class.IsFailure() {
		workflowFailures.Add(ctx, 1, metric.WithAttributes(
			attribute.String(MetricAttrClass, string(info.Class)),
			attribute.String(MetricAttrCategory, info.Category),
			attribute.String(MetricAttrWorkflowType, workflowType),
			attribute.Bool(MetricAttrRoot, root),
		))
	}
}

// RecordActivityFailure counts a failed activity by class and category; cancellations and timeouts are not counted
func RecordActivityFailure(ctx context.Context, activityType string, info Info) {
	if !info.Class.IsFailure() {
		return
	}
	activityFailures.Add(ctx, 1, metric.WithAttributes(
		attribute.String(MetricAttrClass, string(info.Class)),
		attribute.String(MetricAttrCategory, info.Category),
		attribute.String(MetricAttrActivityType, activityType),
	))
}
