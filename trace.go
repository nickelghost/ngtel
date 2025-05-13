package ngtel

import (
	"context"
	"fmt"
	"log/slog"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/api/option"
)

// GetCloudTracer gives us a fully set up tracer for usage in Google Cloud.
// It also gives us a dummy tracer in case of the logging not being enabled.
// The shutdown function should be called when the application is shutting down to ensure all traces are sent.
func GetCloudTracer(
	ctx context.Context, enabled bool, gcpProject string, serviceName string,
) (trace.Tracer, func(), error) {
	if !enabled {
		return noop.NewTracerProvider().Tracer("main"), func() {}, nil
	}

	exporter, err := texporter.New(
		texporter.WithProjectID(gcpProject),
		texporter.WithTraceClientOptions([]option.ClientOption{option.WithTelemetryDisabled()}),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithDetectors(gcp.NewDetector()),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

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

// GetCloudTracePath gives us the path identifier of our current trace, enabling us to connect it in logs for example.
// It returns an empty string if the trace ID is not set.
func GetCloudTracePath(ctx context.Context) string {
	sc := trace.SpanContextFromContext(ctx)
	if sc.HasTraceID() {
		return fmt.Sprintf("projects/%s/traces/%s", GetGCPProject(ctx), sc.TraceID().String())
	}

	return ""
}
