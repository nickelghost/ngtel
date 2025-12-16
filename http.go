package ngtel

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

func RequestMiddleware(handler http.Handler) http.Handler {
	return otelhttp.NewHandler(handler, "request", otelhttp.WithSpanNameFormatter(
		func(operation string, r *http.Request) string {
			return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		},
	))
}

func SetSpanNameMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Pattern != "" {
			span := trace.SpanFromContext(r.Context())
			span.SetName(r.Pattern)
		}

		next.ServeHTTP(w, r)
	})
}
