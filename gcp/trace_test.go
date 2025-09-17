package ngtelgcp_test

import (
	"path/filepath"
	"runtime"
	"testing"

	ngtelgcp "github.com/nickelghost/ngtel/gcp"
	"go.opentelemetry.io/otel/trace"
)

func TestGetTracePath(t *testing.T) {
	testData := []struct {
		traceID   string
		projectID string
		result    string
	}{
		{
			traceID:   "0123456789abcdef0123456789abcdef",
			projectID: "testing-project",
			result:    "projects/testing-project/traces/0123456789abcdef0123456789abcdef",
		},
		{
			traceID:   "12312312312312312312312312312300",
			projectID: "my-project-123asd",
			result:    "projects/my-project-123asd/traces/12312312312312312312312312312300",
		},
	}

	t.Run("project id from env", func(t *testing.T) {
		for _, td := range testData {
			ctx := t.Context()
			traceID, _ := trace.TraceIDFromHex(td.traceID)
			spanID, _ := trace.SpanIDFromHex("0123456789abcdef")
			sc := trace.NewSpanContext(trace.SpanContextConfig{
				TraceID: traceID,
				SpanID:  spanID,
			})
			ctx = trace.ContextWithSpanContext(ctx, sc)

			t.Setenv("GOOGLE_CLOUD_PROJECT", td.projectID)

			tp := ngtelgcp.GetTracePath(ctx)

			if tp != td.result {
				t.Errorf("wrong trace path: %s\ninstead of: %s", tp, td.result)
			}
		}
	})

	t.Run("project id from credentials", func(t *testing.T) {
		_, filename, _, _ := runtime.Caller(0)
		t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", filepath.Dir(filename)+"/test/credentials.json")
		t.Setenv("GOOGLE_CLOUD_PROJECT", "")

		ctx := t.Context()
		traceID, _ := trace.TraceIDFromHex(testData[0].traceID)
		spanID, _ := trace.SpanIDFromHex("0123456789abcdef")
		sc := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: traceID,
			SpanID:  spanID,
		})
		ctx = trace.ContextWithSpanContext(ctx, sc)

		tp := ngtelgcp.GetTracePath(ctx)

		wants := "projects/dummy-project-123/traces/" + traceID.String()

		if tp != wants {
			t.Errorf("wrong trace path: %s\ninstead of: %s", tp, wants)
		}
	})
}
