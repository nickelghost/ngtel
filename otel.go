package ngtel

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var ErrAutoexporterCreationFailed = errors.New("autoexporter creation failed")

// ConfigureOtel configures tracer in automatic mode.
// The shutdown function should be called when the application is shutting down to ensure all traces are sent.
// Configures the batcher and autoprop itself.
func ConfigureOtel(ctx context.Context, samplerName string, samplerRatio float64) (func(), error) {
	otel.SetTextMapPropagator(autoprop.NewTextMapPropagator())

	texporter, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrAutoexporterCreationFailed, err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(texporter),
		sdktrace.WithSampler(getTracesSampler(samplerName, samplerRatio)),
	)

	otel.SetTracerProvider(tp)

	shutdown := func() {
		if err := tp.Shutdown(ctx); err != nil {
			slog.Error("failed to shut down tracer provider", "err", err)
		}
	}

	return shutdown, nil
}
