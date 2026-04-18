package middleware

import (
	"encoding/json"
	"net/http"
)

// validAPIKey is hardcoded for now.
// Stage 07 (Config) will move this to an environment variable.
const validAPIKey = "secret-key-123"

// AuthGuard checks for a valid API key in the X-API-Key request header.
// If missing or wrong → returns 401 Unauthorized and stops the chain.
// If correct → passes the request through to the next handler.
//
// This is applied only to /api/v1/* routes, NOT to /health.
// That's the power of route-level middleware vs global middleware.
func AuthGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// r.Header.Get reads a request header by name (case-insensitive)
		apiKey := r.Header.Get("X-API-Key")

		if apiKey == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "missing X-API-Key header",
			})
			return // IMPORTANT: return here — do NOT call next
		}

		if apiKey != validAPIKey {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "invalid API key",
			})
			return
		}

		// Key is valid — pass the request to the next handler
		next.ServeHTTP(w, r)
	})
}
