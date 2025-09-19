package ngtel

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2/google"
)

// ConfigureTracing gives us tracer in automatic mode.
// The shutdown function should be called when the application is shutting down to ensure all traces are sent.
func ConfigureTracing(ctx context.Context) (func(), error) {
	otel.SetTextMapPropagator(autoprop.NewTextMapPropagator())

	texporter, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(texporter))

	otel.SetTracerProvider(tp)

	shutdown := func() {
		if err := tp.Shutdown(ctx); err != nil {
			slog.Error("failed to shut down tracer provider", "err", err)
		}
	}

	return shutdown, nil
}

// GetGCPTracePath gives us the path identifier of our current trace, enabling us to connect it in logs for example.
// It returns an empty string if the trace ID isn't valid or Google Cloud project ID could not be found.
func GetGCPTracePath(ctx context.Context) string {
	sc := trace.SpanContextFromContext(ctx)
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")

	if projectID == "" {
		creds, _ := google.FindDefaultCredentials(ctx)
		projectID = creds.ProjectID
	}

	if projectID == "" || !sc.TraceID().IsValid() {
		return ""
	}

	tracePath := fmt.Sprintf("projects/%s/traces/%s", projectID, sc.TraceID())

	return tracePath
}
