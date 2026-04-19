// package middleware contains JWT validation middleware.
// This replaces the simple X-API-Key check from previous stages.
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// jwtSecret must match the one used in handlers/auth.go to sign tokens.
// Stage 07 will load this from an environment variable.
var jwtSecret = []byte("jwt-secret-key-change-in-production")

// contextKey avoids key collisions in context (same pattern as stage 03)
type contextKey string

const UserIDKey contextKey = "userID"
const UserEmailKey contextKey = "userEmail"

// JWTAuth validates the Bearer token in the Authorization header.
// If valid — attaches user_id and email to context, calls next.
// If invalid/missing — returns 401.
func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing Authorization header"})
			return
		}

		// Header format must be: "Bearer <token>"
		// strings.Cut splits on the first occurrence of " "
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid Authorization format, use: Bearer <token>"})
			return
		}

		tokenString := parts[1]

		// jwt.Parse validates the token signature and expiry automatically.
		// The keyFunc returns the secret key for verification.
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Verify the signing method is what we expect (HS256)
			// This prevents algorithm confusion attacks
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
			return
		}

		// Extract claims from the valid token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token claims"})
			return
		}

		// Attach user_id and email to context so handlers can use them
		ctx := context.WithValue(r.Context(), UserIDKey, claims["user_id"])
		ctx = context.WithValue(ctx, UserEmailKey, claims["email"])

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID reads the user ID from context (set by JWTAuth middleware).
// Use in handlers: userID := middleware.GetUserID(r.Context())
func GetUserID(ctx context.Context) string {
	id, _ := ctx.Value(UserIDKey).(string)
	return id
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
