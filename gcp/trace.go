package ngtelgcp

import (
	"fmt"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	gcpdetector "go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/api/option"
)

// GetTracingOpts gives us options used for the tracer for Google Cloud.
func GetTracingOpts() ([]resource.Option, []sdktrace.TracerProviderOption, error) {
	exporter, err := texporter.New(
		texporter.WithTraceClientOptions([]option.ClientOption{option.WithTelemetryDisabled()}),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	return []resource.Option{resource.WithDetectors(gcpdetector.NewDetector())},
		[]sdktrace.TracerProviderOption{sdktrace.WithBatcher(exporter)},
		nil
}
