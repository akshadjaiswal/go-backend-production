// package middleware contains HTTP middleware for the application.
//
// This file: RequestLogger — structured request logging.
// Every HTTP request gets logged with method, path, status, duration, and a
// unique request ID. In production (Docker), this JSON goes to stdout and is
// collected by Docker's log driver.
package middleware

import (
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
//
// Problem: http.ResponseWriter doesn't expose the status code after it's been sent.
// We need to log the status AFTER the handler runs, so we intercept WriteHeader().
type responseWriter struct {
	http.ResponseWriter        // embed original — delegates all other methods
	statusCode          int    // captured status code
	written             bool   // track if WriteHeader was called
}

// WriteHeader intercepts the status code before forwarding to the real writer.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.written = true
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the implicit 200 when a handler writes a body without calling WriteHeader.
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

// RequestLogger logs every HTTP request as a structured slog event.
//
// Logged fields:
//   - method, path, status, duration — standard access log fields
//   - request_id — unique per-request ID for correlating logs
//   - remote_addr — caller's IP address
//
// Log level is based on status code:
//   - 5xx → ERROR
//   - 4xx → WARN
//   - rest → INFO
//
// In Docker/production, these become JSON lines in container stdout.
// Log aggregators can then filter: level=ERROR, path=/api/v1/users, etc.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Short unique ID to trace a single request across log lines.
		// Format: last 6 digits of Unix nanoseconds + random 4-digit number
		// e.g. "430802-3421"
		requestID := fmt.Sprintf("%d-%04d", time.Now().UnixNano()%1_000_000, rand.Intn(9999))

		// Clients can use X-Request-ID to reference the request in bug reports.
		w.Header().Set("X-Request-ID", requestID)

		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Run the actual handler
		next.ServeHTTP(wrapped, r)

		// Log after handler returns — we now have the final status code
		duration := time.Since(start)

		logFn := slog.Info
		if wrapped.statusCode >= 500 {
			logFn = slog.Error
		} else if wrapped.statusCode >= 400 {
			logFn = slog.Warn
		}

		logFn("request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", wrapped.statusCode),
			slog.String("duration", duration.String()),
			slog.String("request_id", requestID),
			slog.String("remote_addr", r.RemoteAddr),
		)
	})
}
