// package middleware contains all our custom HTTP middleware.
// Middleware in Go is just a function with this signature:
//   func(http.Handler) http.Handler
// It receives the next handler, wraps it, and returns a new handler.
package middleware

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// contextKey is a custom type for context keys.
// We use a custom type (not plain string) to avoid collisions with
// other packages that might use the same string key in context.
type contextKey string

// RequestIDKey is the key used to store/retrieve the request ID from context.
// Exported so handlers can read it: middleware.RequestIDKey
const RequestIDKey contextKey = "requestID"

// RequestID is our first custom middleware.
// It generates a unique ID for every request and attaches it to the context.
// Why? So you can trace a single request through all your logs.
func RequestID(next http.Handler) http.Handler {
	// http.HandlerFunc converts a function into an http.Handler
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate a simple unique ID: timestamp + random number
		// In production you'd use a proper UUID library (Stage 06)
		id := fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Intn(9999))

		// context.WithValue attaches a value to the request context.
		// The context travels with the request through the entire middleware chain.
		// Any middleware or handler downstream can read this value.
		ctx := context.WithValue(r.Context(), RequestIDKey, id)

		// Also set it as a response header so the client can see it
		w.Header().Set("X-Request-ID", id)

		// r.WithContext returns a new request with the updated context.
		// We pass this new request (not the original) to the next handler.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID is a helper to read the request ID from context in any handler.
// Usage: id := middleware.GetRequestID(r.Context())
func GetRequestID(ctx context.Context) string {
	id, ok := ctx.Value(RequestIDKey).(string)
	if !ok {
		return "unknown"
	}
	return id
}
