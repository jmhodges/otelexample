package main

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
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

	tracer := otel.Tracer("otelexample")
	var traceProviderOpts []sdktrace.TracerProviderOption

	traceExporter, err := stdouttrace.New()
	if err != nil {
		log.Fatalf("unable to create OpenTelemetry stdout tracer: %v", err)
	}

	reqMetricCtx, span := tracer.Start(context.Background(), "loop-request")
	span.SetStatus(codes.Ok, "")
	span.SetAttributes(attribute.Int("did-thing", 10))
	counter.Add(reqMetricCtx, 1)
	span.End()

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
	traceProviderOpts = append(traceProviderOpts, sdktrace.WithResource(resource))
	metricProviderOpts = append(metricProviderOpts, sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(2*time.Second))))
	traceProviderOpts = append(traceProviderOpts, sdktrace.WithBatcher(traceExporter))
	mp := sdkmetric.NewMeterProvider(metricProviderOpts...)
	otel.SetMeterProvider(mp)
	defer mp.Shutdown(context.Background())
	tp := sdktrace.NewTracerProvider(traceProviderOpts...)
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	for {
		// Remember that Contexts that have timed out will not emit new metrics,
		// so if you want timeout metrics to be emitted correctly, you may need
		// to copy the span data to a different Context with a different timeout
		// and use that to send metrics. This collides with some behaviors
		// people like where they get to set one overriding timeout for all of
		// the operations in handling requests and then use it everywhere.
		ctx := context.Background()
		reqMetricCtx := context.Background()
		_, span := tracer.Start(ctx, "loop-request")
		span.SetStatus(codes.Ok, "")
		span.SetAttributes(attribute.Int("did-thing", 20))
		counter.Add(reqMetricCtx, 1)
		span.End()
		time.Sleep(1 * time.Second)
	}
}
