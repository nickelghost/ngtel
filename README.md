# ngtel

Common Go telemetry helpers when using otel and Google Cloud Monitoring.

## Installation

```sh
go get github.com/nickelghost/ngtel
```

## Usage

### Initializing OpenTelemetry

Call `ConfigureOtel` early in your application startup. It configures the tracer provider using the [autoexport](https://pkg.go.dev/go.opentelemetry.io/contrib/exporters/autoexport) and [autoprop](https://pkg.go.dev/go.opentelemetry.io/contrib/propagators/autoprop) packages, so the exporter and propagator are controlled via environment variables (e.g. `OTEL_EXPORTER_OTLP_ENDPOINT`).

The `samplerName` parameter accepts the same values as the `OTEL_TRACES_SAMPLER` env var (`always_on`, `always_off`, `traceidratio`, `parentbased_always_off`, `parentbased_traceidratio`). Any other value falls back to `parentbased_always_on`.

```go
shutdown, err := ngtel.ConfigureOtel(ctx, "parentbased_always_on", 0)
if err != nil {
    log.Fatal(err)
}

defer shutdown(context.Background())
```

### GCP trace correlation in logs

To connect traces with log entries in Google Cloud Logging, set the project ID and pass the trace path as a log field.

`SetProjectID` is optional — if omitted, the library will attempt to auto-detect the project ID from the Application Default Credentials.

```go
ngtel.SetProjectID("my-gcp-project")

// Later, inside a request handler or any function with a traced context:
slog.InfoContext(ctx, "handling request", ngtel.GetGCPLogArgs(ctx)...)
```

`GetGCPLogArgs` returns `[]any{"trace", "<trace-path>"}` when there is a valid trace in the context, or `nil` when there isn't, so it is safe to always spread it into your log calls.

If you need the raw trace path string, use `GetGCPTracePath(ctx)` directly.

### HTTP middleware

Wrap your HTTP handler with `RequestMiddleware` to automatically create spans for incoming requests:

```go
http.ListenAndServe(":8080", ngtel.RequestMiddleware(mux))
```

For routers that expose the matched route pattern via `r.Pattern` (e.g. the standard library's `http.ServeMux` in Go 1.22+), add `SetSpanNameMiddleware` inside individual routes so the span is named after the route pattern instead of the raw URL path:

```go
mux.HandleFunc("GET /users/{id}", ngtel.SetSpanNameMiddleware(func(w http.ResponseWriter, r *http.Request) {
    // span name will be "GET /users/{id}"
}))
```
