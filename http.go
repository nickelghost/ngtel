package ngtel

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

// RequestMiddleware is an HTTP middleware that instruments incoming HTTP requests.
// It creates spans for each request and names them based on the HTTP method and URL path.
func RequestMiddleware(handler http.Handler) http.Handler {
	return otelhttp.NewHandler(handler, "request", otelhttp.WithSpanNameFormatter(
		func(operation string, r *http.Request) string {
			return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		},
	))
}

// SetSpanNameMiddleware sets the span name to the request pattern if available.
// Only effective if used after the router has matched the request.
// This is useful when using routers that support named routes.
func SetSpanNameMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Pattern != "" {
			span := trace.SpanFromContext(r.Context())
			span.SetName(r.Pattern)
		}

		next.ServeHTTP(w, r)
	})
}
