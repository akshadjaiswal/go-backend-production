// package logger sets up the application-wide structured logger using Go's
// built-in `slog` package (available since Go 1.21 — no external dependency needed).
//
// Why structured logging?
// Plain text logs (fmt.Println, log.Printf) are human-readable but machine-unreadable.
// In production, logs go to tools like Datadog, CloudWatch, or Grafana Loki.
// These tools expect JSON — structured fields they can filter, search, and alert on.
//
// Example plain log:     "2026-04-21 10:00:01 GET /users 200 45ms"
// Example structured:    {"time":"2026-04-21T10:00:01Z","level":"INFO","msg":"request","method":"GET","path":"/users","status":200,"duration":"45ms"}
package logger

import (
	"log/slog"
	"os"
)

// Setup initialises the global slog logger based on the environment.
// Call this once at the start of main() before anything else logs.
//
// In dev: DEBUG level + text format (easier to read in terminal)
// In production: INFO level + JSON format (machine-parseable)
func Setup(env string) {
	var handler slog.Handler

	// slog.HandlerOptions lets us configure the minimum log level.
	// Logs below this level are silently dropped.
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug, // default: log everything in dev
	}

	if env == "production" {
		// In production: JSON output, INFO level minimum (skip DEBUG noise)
		opts.Level = slog.LevelInfo
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		// In dev: text output with colours — easier to read in terminal
		// Still structured — just human-friendly format
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	// slog.SetDefault replaces the global logger.
	// After this, slog.Info(), slog.Error(), etc. use our configured handler.
	slog.SetDefault(slog.New(handler))
}

// Info logs an INFO level message with optional key-value pairs.
// Convenience wrapper so callers don't need to import log/slog directly.
func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

// Error logs an ERROR level message with optional key-value pairs.
func Error(msg string, args ...any) {
	slog.Error(msg, args...)
}

// Debug logs a DEBUG level message — only visible in dev environment.
func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

// Warn logs a WARN level message.
func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}
