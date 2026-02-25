package ngtel

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2/google"
)

var (
	projectID     atomic.Pointer[string]
	projectIDOnce sync.Once
)

// SetProjectID sets the Google Cloud project ID to be used for trace paths.
// Ideally, this should be called synchronously during initialization before any trace paths are generated.
func SetProjectID(id string) {
	projectID.Store(&id)

	// Reset the once to allow re-detection of credentials if needed
	projectIDOnce = sync.Once{}
}

// GetGCPTracePath gives us the path identifier of our current trace, enabling us to connect it in logs for example.
// It returns an empty string if the trace ID isn't valid or Google Cloud project ID could not be found.
func GetGCPTracePath(ctx context.Context) string {
	sc := trace.SpanContextFromContext(ctx)
	// IsValid() checks for a non-zero TraceID. trace.TraceID is a [16]byte, so
	// its String() output is always a well-formed 32-char hex string — no
	// additional hex validation is required beyond this zero check.
	if !sc.TraceID().IsValid() {
		return ""
	}

	pID := projectID.Load()

	if pID == nil || *pID == "" {
		projectIDOnce.Do(func() {
			// Check if the project ID was set before once could run, to avoid unnecessary lookups.
			if val := projectID.Load(); val != nil && *val != "" {
				return
			}

			creds, err := google.FindDefaultCredentials(ctx)
			if err != nil {
				slog.Warn("ngtel: could not detect GCP project ID", "err", err)

				return
			}

			projectID.Store(&creds.ProjectID)
		})

		pID = projectID.Load()
	}

	if pID == nil || *pID == "" {
		return ""
	}

	return fmt.Sprintf("projects/%s/traces/%s", *pID, sc.TraceID())
}

func getTracesSampler(name string, ratio float64) sdktrace.Sampler {
	switch strings.ToLower(name) {
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
