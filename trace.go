package ngtel

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// CreateTracer gives us a tracer without an exporter.
// It also gives us a dummy tracer in case of the logging not being enabled.
// The shutdown function should be called when the application is shutting down to ensure all traces are sent.
func CreateTracer(
	ctx context.Context,
	enabled bool,
	serviceName string,
	tpOpts []sdktrace.TracerProviderOption,
	resOpts []resource.Option,
) (trace.Tracer, func(), error) {
	if !enabled {
		return noop.NewTracerProvider().Tracer("main"), func() {}, nil
	}

	res, err := resource.New(ctx, append(
		resOpts,
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(append(
		tpOpts,
		sdktrace.WithResource(res),
	)...)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	otel.SetTracerProvider(tp)

	shutdown := func() {
		if err := tp.Shutdown(ctx); err != nil {
			slog.Error("failed to shut down tracer provider", "err", err)
		}
	}

	return tp.Tracer("main"), shutdown, nil
}
