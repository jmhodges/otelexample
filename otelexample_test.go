package main

import (
	"context"
	"slices"
	"testing"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// TestOtelDroppingContextData is a test that demonstrates how the v1.21.0 and
// earlier OpenTelemetry sdk/metrics package drops data when the Context is
// already canceled. It's a reminder to me (and others) to use a different
// context for recording error states. This was fixed in versions of
// OpenTelemetry later than v1.21.0. See
// https://github.com/open-telemetry/opentelemetry-go/commit/8e756513a630cc0e80c8b65528f27161a87a3cc8
func TestOtelDroppingContextData(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	defer provider.Shutdown(context.Background())

	scopeName := t.Name()
	counterName := "ticks"
	meter := provider.Meter(scopeName)

	tick, err := meter.Int64Counter(counterName)
	if err != nil {
		t.Fatalf("unable to make the ticks counter: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// The meat of the work is adding to the counter with a live Context and
	// then attempting to add to it with a canceled Context.
	tick.Add(ctx, 1)
	cancel()
	tick.Add(ctx, 1)

	rm := new(metricdata.ResourceMetrics)
	err = reader.Collect(context.Background(), rm)
	if err != nil {
		t.Fatalf("failed to collect metrics: %v", err)
	}

	scopeIndex := slices.IndexFunc(rm.ScopeMetrics, func(sm metricdata.ScopeMetrics) bool { return sm.Scope.Name == scopeName })
	if scopeIndex == -1 {
		t.Fatalf("expected to find a ScopeMetrics with the name %s, got none", scopeName)
	}
	scope := rm.ScopeMetrics[scopeIndex]
	metricIndex := slices.IndexFunc(scope.Metrics, func(m metricdata.Metrics) bool { return m.Name == counterName })
	onlyMetric := scope.Metrics[metricIndex]
	onlySum, ok := onlyMetric.Data.(metricdata.Sum[int64])
	if !ok {
		t.Fatalf("expected our ticks Aggregation to be a metricdata.Sum[int64], got %T", onlyMetric.Data)
	}
	datapoint := onlySum.DataPoints[0]
	if datapoint.Value != 1 {
		t.Errorf("expected our ticks Aggregation to have a value of 1 because the second add is dropped since its Context is errored out, got %d", datapoint.Value)
	}

}
