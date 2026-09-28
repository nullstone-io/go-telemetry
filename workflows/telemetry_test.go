package workflows

import (
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// The global provider delegates exactly once, so one reader serves the package; counters are cumulative
var reader = func() *sdkmetric.ManualReader {
	r := sdkmetric.NewManualReader()
	otel.SetMeterProvider(sdkmetric.NewMeterProvider(sdkmetric.WithReader(r)))
	return r
}()

func dataPoints(t *testing.T, name string, match ...attribute.KeyValue) []metricdata.DataPoint[int64] {
	t.Helper()
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatal(err)
	}
	var found []metricdata.DataPoint[int64]
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != name {
				continue
			}
			sum, ok := m.Data.(metricdata.Sum[int64])
			if !ok {
				continue
			}
		points:
			for _, dp := range sum.DataPoints {
				for _, kv := range match {
					if v, ok := dp.Attributes.Value(kv.Key); !ok || v != kv.Value {
						continue points
					}
				}
				found = append(found, dp)
			}
		}
	}
	return found
}

func attrsOf(dp metricdata.DataPoint[int64]) map[string]string {
	result := map[string]string{}
	for _, kv := range dp.Attributes.ToSlice() {
		result[string(kv.Key)] = kv.Value.Emit()
	}
	return result
}

func TestRecordWorkflowCompletion(t *testing.T) {
	ctx := context.Background()
	RecordWorkflowCompletion(ctx, "wf-user", true, User(CategoryTerraformPlan), errors.New("bad hcl"))
	RecordWorkflowCompletion(ctx, "wf-cancelled", true, Cancelled, errors.New("cancelled"))
	RecordWorkflowCompletion(ctx, "wf-ok", false, Info{}, nil)

	failures := dataPoints(t, MetricWorkflowFailures, attribute.String(MetricAttrWorkflowType, "wf-user"))
	if len(failures) != 1 || failures[0].Value != 1 {
		t.Fatalf("expected one failure datapoint of 1, got %+v", failures)
	}
	if got := attrsOf(failures[0]); got[MetricAttrClass] != "user" || got[MetricAttrCategory] != CategoryTerraformPlan || got[MetricAttrRoot] != "true" {
		t.Errorf("failure attributes = %v", got)
	}
	if got := dataPoints(t, MetricWorkflowFailures, attribute.String(MetricAttrWorkflowType, "wf-cancelled")); len(got) != 0 {
		t.Error("a cancellation must not count as a failure")
	}
	for wf, status := range map[string]string{"wf-user": CompletionFailed, "wf-cancelled": CompletionCancelled, "wf-ok": CompletionSucceeded} {
		completions := dataPoints(t, MetricWorkflowCompletions, attribute.String(MetricAttrWorkflowType, wf))
		if len(completions) != 1 || attrsOf(completions[0])[MetricAttrStatus] != status {
			t.Errorf("%s: expected one completion with status %q, got %+v", wf, status, completions)
		}
	}
}

func TestRecordActivityFailure(t *testing.T) {
	ctx := context.Background()
	RecordActivityFailure(ctx, "act-user", User(CategoryDockerBuild))
	RecordActivityFailure(ctx, "act-cancelled", Cancelled)

	failures := dataPoints(t, MetricActivityFailures, attribute.String(MetricAttrActivityType, "act-user"))
	if len(failures) != 1 || attrsOf(failures[0])[MetricAttrCategory] != CategoryDockerBuild {
		t.Errorf("expected one docker-build activity failure, got %+v", failures)
	}
	if got := dataPoints(t, MetricActivityFailures, attribute.String(MetricAttrActivityType, "act-cancelled")); len(got) != 0 {
		t.Error("a cancellation must not count as a failure")
	}
}
