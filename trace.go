package ngtel

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2/google"
)

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

func getTracesSampler() sdktrace.Sampler {
	samplerName := os.Getenv("OTEL_TRACES_SAMPLER")
	ratio := 1.0

	if arg := os.Getenv("OTEL_TRACES_SAMPLER_ARG"); arg != "" {
		if r, err := strconv.ParseFloat(arg, 64); err == nil {
			ratio = r
		}
	}

	switch strings.ToLower(samplerName) {
	case "always_on":
		return sdktrace.AlwaysSample()
	case "always_off":
		return sdktrace.NeverSample()
	case "traceidratio":
		return sdktrace.TraceIDRatioBased(ratio)
	case "parentbased_always_off":
		return sdktrace.ParentBased(sdktrace.NeverSample())
	case "parentbased_traceidratio":
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))
	default:
		return sdktrace.ParentBased(sdktrace.AlwaysSample())
	}
}
