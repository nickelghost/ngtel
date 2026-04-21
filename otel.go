package ngtel

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var ErrExporterCreationFailed = errors.New("exporter creation failed")

// ConfigureOtel configures tracer in automatic mode or using GRPC creds if provided.
// The shutdown function should be called when the application is shutting down to ensure all traces are sent.
// Configures the batcher and autoprop itself.
func ConfigureOtel(
	ctx context.Context, samplerName string, samplerRatio float64, creds *credentials.PerRPCCredentials,
) (func(context.Context), error) {
	otel.SetTextMapPropagator(autoprop.NewTextMapPropagator())

	var (
		e   sdktrace.SpanExporter
		err error
	)

	if creds != nil {
		e, err = otlptracegrpc.New(ctx, otlptracegrpc.WithDialOption(grpc.WithPerRPCCredentials(*creds)))
	} else {
		e, err = autoexport.NewSpanExporter(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrExporterCreationFailed, err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(e),
		sdktrace.WithSampler(getTracesSampler(samplerName, samplerRatio)),
	)

	otel.SetTracerProvider(tp)

	shutdown := func(ctx context.Context) {
		if err := tp.Shutdown(ctx); err != nil {
			slog.Error("failed to shut down tracer provider", "err", err)
		}
	}

	return shutdown, nil
}
