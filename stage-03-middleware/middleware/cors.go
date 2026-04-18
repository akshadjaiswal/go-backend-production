package middleware

import "net/http"

// CORS adds Cross-Origin Resource Sharing headers to every response.
// Without this, browsers block requests from a different domain (e.g. your
// React frontend on localhost:3000 calling your API on localhost:8080).
//
// These headers tell the browser: "yes, this API accepts requests from other origins."
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from any origin (*).
		// In production, replace * with your specific frontend domain.
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Which HTTP methods are allowed cross-origin
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// Which headers the client is allowed to send
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, X-Request-ID")

		// Preflight request: browsers send OPTIONS before the real request
		// to check if CORS is allowed. We respond 200 and return early.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
