// Package testhelpers provides shared utilities used across all test files in stage-09.
//
// Why a separate package?
// Test code in Go lives in _test.go files, but those files are package-scoped.
// If both handlers/auth_test.go and handlers/users_test.go need "set up a test DB"
// and "create a valid JWT", we'd duplicate that logic in both files.
//
// A separate non-test package (testhelpers) lets any _test.go file import and reuse it.
// It is NOT a _test.go file itself, so it's importable by other packages' test files.
// The Go toolchain still excludes it from the production binary because no non-test
// code imports it.
package testhelpers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver — blank import registers it
	"github.com/joho/godotenv"

	"github.com/akshadjaiswal/go-backend-production/stage-09-testing/config"
)

// SetupTestDB connects to the test database and runs the users migration.
//
// It reads TEST_DATABASE_URL from the .env.test file (or from env vars if already set).
// If the var is missing, the test fails immediately with a clear message.
//
// t.Helper() marks this as a helper function — if it fails, Go's test output
// will point to the caller's line number, not to this function. Much easier to debug.
func SetupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	// go test runs each package from that package's own directory.
	// Try both common relative paths so this works regardless of which package calls it.
	loadEnvTest()


	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Fatal("TEST_DATABASE_URL not set — create a .env.test file or set the env var")
	}

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to open test DB: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping test DB: %v\n\nMake sure you ran: createdb go_backend_production_stage09_test", err)
	}

	// Run migration — CREATE TABLE IF NOT EXISTS is idempotent, safe to run every time.
	// This means even if the table already exists from a previous test run, it won't fail.
	_, err = db.Exec(`
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

		CREATE TABLE IF NOT EXISTS users (
			id            UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
			name          TEXT        NOT NULL,
			email         TEXT        NOT NULL UNIQUE,
			password_hash TEXT        NOT NULL DEFAULT '',
			created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		t.Fatalf("failed to run migration: %v", err)
	}

	return db
}

// CleanupDB wipes the users table between tests.
//
// Why TRUNCATE instead of DROP TABLE?
// - DROP TABLE + re-create is slow and risky
// - TRUNCATE is fast — it resets the table to empty in one operation
// - CASCADE handles foreign keys (if we add them in later stages)
//
// Call this at the START of each test, not the end. That way:
// - If a test fails halfway, the DB is still dirty
// - But the NEXT test cleans up at its start — always a clean slate
// - You can inspect the DB after a failed test to debug it
//
// testing.TB is an interface satisfied by both *testing.T and *testing.B
// (benchmark tests). Using TB makes CleanupDB usable in both contexts.
func CleanupDB(t testing.TB, db *sqlx.DB) {
	t.Helper()
	_, err := db.Exec("TRUNCATE users CASCADE")
	if err != nil {
		t.Fatalf("failed to clean up test DB: %v", err)
	}
}

// MakeTestConfig returns a *config.Config wired for tests.
//
// Key differences from production config:
// - Points at the test database (read from TEST_DATABASE_URL)
// - Uses a fixed, known JWT secret ("test-secret") so tests can create tokens
// - Short JWT expiry (1 hour) — doesn't matter for tests but is realistic
// - ENV is set to "test" (suppresses slog output in tests)
func MakeTestConfig() *config.Config {
	loadEnvTest()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://localhost/go_backend_production_stage09_test?sslmode=disable"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "test-secret-key-for-stage09"
	}

	return &config.Config{
		Port:           "8080",
		Env:            "test",
		DatabaseURL:    dbURL,
		JWTSecret:      jwtSecret,
		JWTExpiryHours: 1,
	}
}

// MakeAuthToken generates a valid JWT string for testing authenticated endpoints.
//
// Instead of going through the full register+login HTTP cycle (which would require
// a running server), we build the token directly using the same JWT library and
// the same signing logic as handlers/auth.go's generateToken method.
//
// This is fine because:
// - We control the secret (it's MakeTestConfig().JWTSecret)
// - The JWT middleware only checks signature validity — it doesn't care how the token was made
// - Integration tests for Login itself test the full HTTP path separately
func MakeAuthToken(t *testing.T, cfg *config.Config) string {
	t.Helper()

	claims := jwt.MapClaims{
		"user_id": "00000000-0000-0000-0000-000000000001", // fake but valid UUID4 shape
		"email":   "testuser@example.com",
		"name":    "Test User",
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		t.Fatalf("failed to generate test auth token: %v", err)
	}
	return signed
}

// NewRequest builds an *http.Request suitable for use with httptest.NewRecorder().
//
// What httptest.NewRequest does vs http.NewRequest:
// - Both create a *http.Request
// - httptest.NewRequest sets the RequestURI field (required for some handlers)
// - Neither sends anything over a real network — it's just a struct in memory
//
// body: pass nil for requests with no body (GET, DELETE)
//       pass a struct for POST/PUT — it gets JSON-encoded automatically
//
// token: pass "" for unauthenticated requests
//        pass a JWT string for authenticated requests — it gets added as "Bearer <token>"
func NewRequest(t *testing.T, method, path string, body any, token string) *http.Request {
	t.Helper()

	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
	}

	// httptest.NewRequest panics on invalid method, which is fine for tests
	req := httptest.NewRequest(method, path, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req
}

// loadEnvTest loads .env.test from common locations.
//
// go test runs each package from its own directory, so the relative path to .env.test
// depends on how deep the package is. We try both:
//   - ".env.test"    → works when running from stage-09-testing/ root
//   - "../.env.test" → works when running from a sub-package like handlers/ or middleware/
//
// godotenv.Load only sets vars that are NOT already set (env vars always win),
// so calling it twice is safe.
func loadEnvTest() {
	_ = godotenv.Load(".env.test")
	_ = godotenv.Load("../.env.test")
}
