package main

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func main() {
	var metricProviderOpts []sdkmetric.Option

	metricExporter, err := stdoutmetric.New()
	if err != nil {
		log.Fatalf("unable to create OpenTelemetry stdout metric exporter: %v", err)
	}

	meter := otel.Meter("otelexample")
	counter, err := meter.Int64Counter("ticks")
	if err != nil {
		log.Fatalf("failed to create counter metric: %v", err)
	}
	counter.Add(context.Background(), 100)

	ctx := context.Background()
	resource, err := resource.New(ctx,
		// Keep the default detectors
		resource.WithTelemetrySDK(),
		// Add your own custom attributes to identify your application
		resource.WithAttributes(
			semconv.ServiceName("otelexample"),
		),
	)
	if err != nil {
		log.Fatalf("failed to create OpenTelemetry resource for otelexample service: %v", err)
	}

	metricProviderOpts = append(metricProviderOpts, sdkmetric.WithResource(resource))
	metricProviderOpts = append(metricProviderOpts, sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(2*time.Second))))
	mp := sdkmetric.NewMeterProvider(metricProviderOpts...)
	otel.SetMeterProvider(mp)
	defer mp.Shutdown(context.Background())

	for {
		counter.Add(context.Background(), 1)
		time.Sleep(1 * time.Second)
	}
}

// var tracerProviderOpts []sdktrace.TracerProviderOption[]
// traceExporter, err := stdouttrace.New()
// if err != nil {
// 	log.Fatalf("unable to create OpenTelemetry stdout tracer: %v", err)
// }
// traceProviderOpts = append(traceProviderOpts, sdktrace.WithSyncer(traceExporter))
