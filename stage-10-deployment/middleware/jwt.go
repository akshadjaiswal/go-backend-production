// JWT authentication middleware.
//
// Protects routes under /api/v1/* — any request without a valid Bearer token
// gets a 401 Unauthorized response before the handler ever runs.
//
// Flow:
//   1. Read Authorization header
//   2. Verify it's "Bearer <token>"
//   3. Parse and validate the JWT (signature + expiry)
//   4. Extract claims (user_id, email) and attach to request context
//   5. Call next handler — which reads the claims with GetUserID()
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/akshadjaiswal/go-backend-production/stage-10-deployment/config"
)

// contextKey is a private type for context keys.
// Using a named type prevents collisions with other packages that also use string keys.
type contextKey string

const UserIDKey contextKey = "userID"
const UserEmailKey contextKey = "userEmail"

// JWTMiddleware holds the config so it can read JWTSecret from config (not hardcoded).
type JWTMiddleware struct {
	cfg *config.Config
}

// NewJWTAuth creates a JWTMiddleware. Usage in routes:
//
//	r.Use(middleware.NewJWTAuth(cfg).Handler)
func NewJWTAuth(cfg *config.Config) *JWTMiddleware {
	return &JWTMiddleware{cfg: cfg}
}

// Handler is the actual middleware function.
// It runs before every protected route handler.
func (m *JWTMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing Authorization header"})
			return
		}

		// Expected format: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid Authorization format, use: Bearer <token>"})
			return
		}

		// jwt.Parse validates the signature using our secret key.
		// The key function is called for each token to provide the verification key.
		// We also check the signing method to prevent algorithm confusion attacks
		// (an attacker could send a "none" algorithm token otherwise).
		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(m.cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token claims"})
			return
		}

		// Attach user info to the request context so handlers can read it.
		// context.WithValue creates a new context with the key-value added.
		ctx := context.WithValue(r.Context(), UserIDKey, claims["user_id"])
		ctx = context.WithValue(ctx, UserEmailKey, claims["email"])
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID extracts the authenticated user's ID from the request context.
// Call this inside any handler that runs behind JWTMiddleware.
func GetUserID(ctx context.Context) string {
	id, _ := ctx.Value(UserIDKey).(string)
	return id
}

// writeJSON is a private helper — writes JSON with the given status code.
// Defined here (not in handlers) so middleware can send error responses
// without importing the handlers package (which would create a cycle).
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
