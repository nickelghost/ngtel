package ngtel_test

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/nickelghost/ngtel"
	"go.opentelemetry.io/otel/trace"
)

func TestGetGCPTracePath(t *testing.T) { //nolint:paralleltest
	tests := []struct {
		name      string
		traceID   string
		projectID string
		want      string
	}{
		{
			name:      "example 1",
			traceID:   "0123456789abcdef0123456789abcdef",
			projectID: "testing-project",
			want:      "projects/testing-project/traces/0123456789abcdef0123456789abcdef",
		},
		{
			name:      "example 2",
			traceID:   "12312312312312312312312312312300",
			projectID: "my-project-123asd",
			want:      "projects/my-project-123asd/traces/12312312312312312312312312312300",
		},
	}

	for _, tt := range tests { //nolint:paralleltest
		t.Run(tt.name, func(t *testing.T) {
			ngtel.SetProjectID(tt.projectID)

			ctx := ctxWithTrace(t, tt.traceID)

			if got := ngtel.GetGCPTracePath(ctx); got != tt.want {
				t.Errorf(`got "%v" wanted "%v"`, got, tt.want)
			}
		})
	}
}

func TestGetGCPTracePath_CredentialsDetection(t *testing.T) {
	ngtel.SetProjectID("")

	_, filename, _, _ := runtime.Caller(0)
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", filepath.Join(filepath.Dir(filename), "test", "credentials.json"))
	t.Setenv("GOOGLE_CLOUD_PROJECT", "")

	const traceID = "0123456789abcdef0123456789abcdef"

	ctx := ctxWithTrace(t, traceID)

	got := ngtel.GetGCPTracePath(ctx)
	want := "projects/dummy-project-123/traces/" + traceID

	if got != want {
		t.Errorf(`got "%v" wanted "%v"`, got, want)
	}
}

func TestGetGCPTracePath_InvalidTraceID(t *testing.T) { //nolint:paralleltest
	ngtel.SetProjectID("testing-project")

	ctx := ctxWithTrace(t, "invalid-trace-id")

	if got := ngtel.GetGCPTracePath(ctx); got != "" {
		t.Errorf(`got "%v" wanted empty string`, got)
	}
}

func TestGetGCPTracePath_NoProjectID(t *testing.T) { //nolint:paralleltest
	ngtel.SetProjectID("")

	ctx := ctxWithTrace(t, "0123456789abcdef0123456789abcdef")

	if got := ngtel.GetGCPTracePath(ctx); got != "" {
		t.Errorf(`got "%v" wanted empty string`, got)
	}
}

func ctxWithTrace(t *testing.T, traceHex string) context.Context {
	t.Helper()

	traceID, _ := trace.TraceIDFromHex(traceHex)
	spanID, _ := trace.SpanIDFromHex("0123456789abcdef")
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID,
		SpanID:  spanID,
	})

	return trace.ContextWithSpanContext(t.Context(), sc)
}
