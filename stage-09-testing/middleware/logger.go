package middleware

import (
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
// We need this because the standard ResponseWriter doesn't expose the status
// after it's been written — but we need it to log it after the handler runs.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.written = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

// RequestLogger is our structured request logging middleware.
// It replaces Chi's built-in middleware.Logger with JSON-structured output.
//
// For every request it logs:
//   - method, path, status code, duration
//   - a unique request_id so you can trace one request across all logs
//   - remote address of the caller
//
// The request_id is also attached to the response header so clients can reference it.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Generate a short unique request ID for tracing this request.
		// Format: timestamp (last 6 digits) + random 4-digit number
		// e.g. "430802-3421"
		requestID := fmt.Sprintf("%d-%04d", time.Now().UnixNano()%1_000_000, rand.Intn(9999))

		// Attach request ID to response header — useful for client-side debugging
		w.Header().Set("X-Request-ID", requestID)

		// Wrap the ResponseWriter so we can read the status code after the handler
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call the next handler — actual work happens here
		next.ServeHTTP(wrapped, r)

		// After the handler returns, log the completed request as structured JSON.
		// slog.With() creates a logger with pre-set fields — all fields below
		// will appear in every log line produced by this logger.
		duration := time.Since(start)

		// Choose log level based on status code:
		// 5xx → ERROR, 4xx → WARN, rest → INFO
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
