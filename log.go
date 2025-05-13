// Package ngtel is a small personal library of telemetry helpers
package ngtel

import (
	"context"
	"io"
	"log/slog"
	"strings"
)

// GetLogLevel determines the slog log level based on the provided string.
// It returns slog.LevelInfo if the provided string does not match any known level.
// The function is case-insensitive and supports the following levels: debug, info, warn, error.
func GetLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	}

	return slog.LevelInfo
}

// SetUpLogger sets the default logger as that of the chosen format.
// It supports JSON, Google Cloud, and Text formats.
func SetUpLogger(w io.Writer, format string, lvl slog.Level) {
	opts := &slog.HandlerOptions{Level: lvl, AddSource: true}

	switch strings.ToLower(format) {
	case "json":
		slog.SetDefault(slog.New(slog.NewJSONHandler(w, opts)))
	case "google_cloud":
		slog.SetDefault(slog.New(NewCloudLoggingHandler(w, opts)))
	default:
		slog.SetDefault(slog.New(slog.NewTextHandler(w, opts)))
	}
}

// CloudLoggingHandler produces (s)logs that conform to Google Cloud's standards and supports its dedicated fields.
type CloudLoggingHandler struct{ handler slog.Handler }

// NewCloudLoggingHandler sets up the Google Cloud handler.
func NewCloudLoggingHandler(w io.Writer, opts *slog.HandlerOptions) *CloudLoggingHandler {
	opts.ReplaceAttr = func(_ []string, a slog.Attr) slog.Attr {
		switch a.Key {
		case slog.MessageKey:
			a.Key = "message"
		case slog.SourceKey:
			a.Key = "logging.googleapis.com/sourceLocation"
		case slog.LevelKey:
			a.Key = "severity"

			level, _ := a.Value.Any().(slog.Level)
			if level == slog.Level(12) { //nolint:mnd
				a.Value = slog.StringValue("CRITICAL")
			}
		case "trace":
			a.Key = "logging.googleapis.com/trace"
		}

		return a
	}

	return &CloudLoggingHandler{handler: slog.NewJSONHandler(w, opts)}
}

func (h *CloudLoggingHandler) Enabled(ctx context.Context, level slog.Level) bool { //nolint:revive
	return h.handler.Enabled(ctx, level)
}

func (h *CloudLoggingHandler) Handle(ctx context.Context, rec slog.Record) error { //nolint:revive
	return h.handler.Handle(ctx, rec) //nolint:wrapcheck
}

func (h *CloudLoggingHandler) WithAttrs(attrs []slog.Attr) slog.Handler { //nolint:revive
	return &CloudLoggingHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *CloudLoggingHandler) WithGroup(name string) slog.Handler { //nolint:revive
	return &CloudLoggingHandler{handler: h.handler.WithGroup(name)}
}
