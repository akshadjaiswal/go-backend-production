// package logger sets up the application-wide structured logger using Go's
// built-in `slog` package (available since Go 1.21 — no external dependency needed).
//
// Why structured logging matters in Docker/production:
//   - Container logs go to stdout and are collected by Docker, Kubernetes, or a log aggregator
//   - Log aggregators (Datadog, CloudWatch, Loki) expect JSON — so they can index and filter fields
//   - In dev, plain text is easier to read in the terminal
//
// Same binary, different format depending on ENV variable:
//   ENV=dev         → text format (human-friendly terminal output)
//   ENV=production  → JSON format (machine-parseable, what docker compose will show)
package logger

import (
	"log/slog"
	"os"
)

// Setup initialises the global slog logger.
// Call this once in main(), before anything else logs.
//
// After this call, slog.Info(), slog.Error(), slog.Debug() etc. work everywhere
// in the app without importing this package again.
func Setup(env string) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug, // log everything in dev
	}

	if env == "production" {
		// JSON output — one JSON object per line, easy to parse by log tools
		// Example: {"time":"2026-04-23T10:00:01Z","level":"INFO","msg":"request","method":"GET","path":"/health","status":200}
		opts.Level = slog.LevelInfo // skip DEBUG noise in production
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		// Text output — human-friendly, coloured in most terminals
		// Example: time=2026-04-23T10:00:01Z level=INFO msg=request method=GET path=/health status=200
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	// Replace the global logger — all slog.* calls from here on use this handler
	slog.SetDefault(slog.New(handler))
}

// Convenience wrappers so callers don't need to import log/slog directly.

func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

func Error(msg string, args ...any) {
	slog.Error(msg, args...)
}

func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}
