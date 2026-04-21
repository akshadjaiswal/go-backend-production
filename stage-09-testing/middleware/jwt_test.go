package middleware_test

// Black-box test — package middleware_test instead of middleware.
// We test the public Handler method exactly as external code uses it.

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/akshadjaiswal/go-backend-production/stage-09-testing/config"
	"github.com/akshadjaiswal/go-backend-production/stage-09-testing/middleware"
)

// testConfig returns a minimal config for JWT middleware tests.
// We don't need a DB URL here — the JWT middleware doesn't touch the database.
func testConfig() *config.Config {
	return &config.Config{
		JWTSecret:      "test-secret-key",
		JWTExpiryHours: 1,
		Env:            "test",
	}
}

// makeToken creates a signed JWT string for testing.
// expired=true creates a token that is already past its expiry time.
func makeToken(secret string, expired bool) string {
	expiry := time.Now().Add(1 * time.Hour)
	if expired {
		expiry = time.Now().Add(-1 * time.Hour) // 1 hour in the past
	}

	claims := jwt.MapClaims{
		"user_id": "abc-123",
		"email":   "test@example.com",
		"exp":     expiry.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(secret))
	return signed
}

// TestJWTMiddleware_Handler tests the JWT middleware using httptest.
//
// How httptest works for middleware testing:
//
//   Without a real server:
//   1. Create a fake response recorder (httptest.NewRecorder)
//      — it implements http.ResponseWriter but writes to memory, not a socket
//   2. Create a fake request (httptest.NewRequest)
//      — it's a real *http.Request struct, just not from a network
//   3. Wrap a simple "next" handler in our middleware
//   4. Call ServeHTTP with the fake recorder + request
//   5. Check the recorder's Code and Body
//
// The middleware doesn't know or care that there's no real network — it just
// sees an http.ResponseWriter and an *http.Request.
func TestJWTMiddleware_Handler(t *testing.T) {
	cfg := testConfig()

	tests := []struct {
		name           string
		authHeader     string       // what to put in Authorization header (empty = no header)
		expectedStatus int
		expectedBody   string       // substring to check in response body
	}{
		{
			name:           "no Authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "missing Authorization header",
		},
		{
			name:           "wrong format — no Bearer prefix",
			authHeader:     "Token abc123",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid Authorization format",
		},
		{
			name:           "Bearer prefix but empty token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid or expired token",
		},
		{
			name:           "completely invalid token string",
			authHeader:     "Bearer not.a.real.jwt",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid or expired token",
		},
		{
			name:           "valid token signed with wrong secret",
			authHeader:     "Bearer " + makeToken("wrong-secret", false),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid or expired token",
		},
		{
			name:           "expired token",
			authHeader:     "Bearer " + makeToken(cfg.JWTSecret, true),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid or expired token",
		},
		{
			name:           "valid token — next handler is called",
			authHeader:     "Bearer " + makeToken(cfg.JWTSecret, false),
			expectedStatus: http.StatusOK, // our "next" handler returns 200
			expectedBody:   "reached",     // our "next" handler writes "reached"
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// "next" is what the middleware calls if auth succeeds.
			// A simple handler that writes 200 + "reached" so we know it was called.
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"message":"reached"}`))
			})

			// Wrap "next" with the JWT middleware
			jwtMiddleware := middleware.NewJWTAuth(cfg)
			handler := jwtMiddleware.Handler(next)

			// Build the fake request
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			// Build the fake response recorder
			// Think of it as a bytes.Buffer + status code field
			rec := httptest.NewRecorder()

			// Call the handler — this runs synchronously, no goroutines needed
			handler.ServeHTTP(rec, req)

			// Check status code
			assert.Equal(t, tc.expectedStatus, rec.Code,
				"unexpected status code for case: %s", tc.name)

			// Check body contains the expected string
			// We use Contains (not Equal) because the body may have extra JSON fields
			assert.Contains(t, rec.Body.String(), tc.expectedBody,
				"unexpected body for case: %s", tc.name)
		})
	}
}

// TestJWTMiddleware_ContextValues verifies that after successful auth,
// the middleware puts the user_id and email into the request context.
//
// This is important because handlers like ListUsers rely on the context having
// userID to know who made the request.
func TestJWTMiddleware_ContextValues(t *testing.T) {
	cfg := testConfig()

	// Create a token with specific claims
	claims := jwt.MapClaims{
		"user_id": "test-user-123",
		"email":   "context@example.com",
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.JWTSecret))
	require.NoError(t, err)

	// A "next" handler that reads from context and asserts on it
	var capturedUserID, capturedEmail string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = middleware.GetUserID(r.Context())
		// We can't easily access userEmail from outside the package (it uses a private key type)
		// but GetUserID covers the main case
		w.WriteHeader(http.StatusOK)
	})

	jwtMiddleware := middleware.NewJWTAuth(cfg)
	handler := jwtMiddleware.Handler(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test-user-123", capturedUserID,
		"middleware should store user_id in context, accessible via GetUserID")
	_ = capturedEmail // not testing this directly — GetUserID already confirms context works
}
