An example of OpenTelemetry demonstrating that you can set the `MeterProvider` with `otel.SetMeterProvider`
after calling `otel.Meter` and get the metrics emitted as needed by the new
settings. The same is true (but not shown here) for `otel.Tracer` and
`TracerProvider`.

One limit (also demonstrated here if you squint at the output) is that any
changes to metric values (see the `counter.Add(..., 100)` call before the
SetProvider that never makes it out) made before setting `MeterProvider` will be
dropped. Values changed after calling `SetMeterProvider` will be emitted as
expected.
