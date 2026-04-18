package middleware

import (
	"fmt"
	"net/http"
	"time"
)

// responseWriter is a custom wrapper around http.ResponseWriter.
// The problem: standard http.ResponseWriter doesn't let you read the status code
// after it's been written. We need the status code to log it.
// Solution: wrap it, intercept WriteHeader(), store the code ourselves.
type responseWriter struct {
	http.ResponseWriter        // embed the original — we get all its methods for free
	statusCode          int    // we'll capture the status code here
	written             bool   // track if WriteHeader was called
}

// WriteHeader intercepts the status code before passing it to the real writer.
// This is called by handlers when they do w.WriteHeader(http.StatusOK).
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.written = true
	rw.ResponseWriter.WriteHeader(code) // still call the real one
}

// Write intercepts body writes. If no status was set yet, default to 200.
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

// Logger logs every request: method, path, status code, duration, request ID.
// It wraps the ResponseWriter so it can capture the status code AFTER the handler runs.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now() // record when request arrived

		// Wrap the response writer so we can capture the status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // default
		}

		// Call the next handler — this is where the actual work happens
		next.ServeHTTP(wrapped, r)

		// After the handler returns, we log everything
		duration := time.Since(start)
		requestID := GetRequestID(r.Context())

		fmt.Printf("[%s] %s %s → %d (%s) | req_id=%s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration,
			requestID,
		)
	})
}
