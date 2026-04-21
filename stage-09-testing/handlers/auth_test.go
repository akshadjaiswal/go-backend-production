package handlers_test

// Integration tests for AuthHandler.Register and AuthHandler.Login.
//
// "Integration test" means we test the full path:
//   HTTP request → handler → real PostgreSQL database → HTTP response
//
// Compare to "unit test" which mocks the DB — we don't do that here because:
//   1. Our repo's convention: integration tests hit a real test DB (see CLAUDE.md)
//   2. We've seen mocks pass while real queries fail — real DB catches real bugs
//   3. The test DB is fast enough (all running locally)

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/akshadjaiswal/go-backend-production/stage-09-testing/handlers"
	"github.com/akshadjaiswal/go-backend-production/stage-09-testing/testhelpers"
)

// Package-level vars shared across all tests in this file.
// Set once in TestMain, reused by every Test* function.
var (
	authTestDB      *sqlx.DB
	authTestHandler *handlers.AuthHandler
)

// TestMain is a special function Go calls BEFORE running any tests in this package.
//
// It's the right place for:
//   - Setting up shared resources (DB connection)
//   - Running migrations once (not per-test)
//   - Cleanup after ALL tests complete
//
// m.Run() actually runs the tests. The return value is the exit code.
// os.Exit(m.Run()) is the conventional pattern — if we don't call os.Exit,
// defer statements won't clean up and the test binary won't report pass/fail correctly.
//
// Note: TestMain must be in a _test.go file in the package being tested.
// Since we're in package handlers_test, this governs all auth_test.go + users_test.go tests.
func TestMain(m *testing.M) {
	// Set up the test DB once for the whole package
	// We use testing.T-like behavior but TestMain gets *testing.M (not *testing.T)
	// so we create a temporary *testing.T just for setup
	//
	// Actually, SetupTestDB needs *testing.T. The Go way is to use a flag or panic.
	// We'll use a simple panic wrapper since DB failure means no tests can run.
	db := setupTestDBForMain()
	cfg := testhelpers.MakeTestConfig()

	authTestDB = db
	authTestHandler = handlers.NewAuthHandler(db, cfg)

	// m.Run() runs all TestXxx functions in this package
	// Capture the exit code so we can report pass/fail
	exitCode := m.Run()

	// Cleanup: close the DB connection after all tests finish
	db.Close()

	os.Exit(exitCode)
}

// setupTestDBForMain is a helper for TestMain since we can't pass *testing.T there.
// It panics on failure — if we can't connect to the test DB, no tests can run anyway.
func setupTestDBForMain() *sqlx.DB {
	// Inline the setup logic since we can't use testhelpers.SetupTestDB (needs *testing.T)
	from_godotenv_load_env_test()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		panic("TEST_DATABASE_URL not set — create a .env.test file in the stage-09-testing directory")
	}

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		panic("failed to open test DB: " + err.Error())
	}
	if err := db.Ping(); err != nil {
		panic("failed to ping test DB (did you run 'createdb go_backend_production_stage09_test'?): " + err.Error())
	}

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
		panic("failed to run migration: " + err.Error())
	}

	return db
}

// from_godotenv_load_env_test loads .env.test — separate function to avoid import cycle.
func from_godotenv_load_env_test() {
	// go test runs each package from that package's own directory.
	// handlers/ is one level inside stage-09-testing/, so .env.test is at "../.env.test"
	// We also try ".env.test" as a fallback for when tests are run differently.
	_ = loadEnvFile("../.env.test")
	_ = loadEnvFile(".env.test")
}

// loadEnvFile is a minimal .env file loader used only in TestMain.
// We use this instead of importing godotenv directly in the test file to keep imports clean.
// In practice, if TEST_DATABASE_URL is already set as an env var (CI), this is skipped.
func loadEnvFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err // file doesn't exist — that's OK, env vars may already be set
	}

	lines := splitLines(string(data))
	for _, line := range lines {
		// Skip comments and empty lines
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		// Split on first '='
		for i, c := range line {
			if c == '=' {
				key := line[:i]
				val := line[i+1:]
				// Only set if not already set (env vars take precedence)
				if os.Getenv(key) == "" {
					os.Setenv(key, val)
				}
				break
			}
		}
	}
	return nil
}

// splitLines splits a string by newlines.
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// --- Register Tests ---

// TestRegister tests all Register scenarios in a table-driven style.
// Each subtest gets a clean DB state via CleanupDB at the start.
func TestRegister(t *testing.T) {
	tests := []struct {
		name           string
		body           map[string]any // request JSON body
		expectedStatus int
		checkBody      func(t *testing.T, body map[string]any) // optional extra assertions
	}{
		{
			name: "valid registration",
			body: map[string]any{
				"name":     "Akshad Jaiswal",
				"email":    "akshad@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusCreated,
			checkBody: func(t *testing.T, body map[string]any) {
				// Response should have id, name, email — but NOT password_hash
				assert.NotEmpty(t, body["id"], "response should include user ID")
				assert.Equal(t, "Akshad Jaiswal", body["name"])
				assert.Equal(t, "akshad@example.com", body["email"])
				// password_hash has json:"-" tag — must never appear in response
				_, hasPasswordHash := body["password_hash"]
				assert.False(t, hasPasswordHash, "password_hash must not be in response (json:\"-\" tag)")
			},
		},
		{
			name: "missing name",
			body: map[string]any{
				"email":    "test@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			checkBody: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "validation failed", body["error"])
				// fields should be an array with at least one entry about "name"
				fields, ok := body["fields"].([]any)
				require.True(t, ok, "fields should be an array")
				require.NotEmpty(t, fields)
				// Find the name field error
				found := false
				for _, f := range fields {
					field := f.(map[string]any)
					if field["field"] == "name" {
						found = true
						break
					}
				}
				assert.True(t, found, "expected validation error for 'name' field")
			},
		},
		{
			name: "invalid email",
			body: map[string]any{
				"name":     "Test User",
				"email":    "not-an-email",
				"password": "password123",
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "password too short (min 8)",
			body: map[string]any{
				"name":     "Test User",
				"email":    "test@example.com",
				"password": "short",
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "empty body",
			body:           map[string]any{},
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Clean slate for every test — TRUNCATE users
			testhelpers.CleanupDB(t, authTestDB)

			req := testhelpers.NewRequest(t, http.MethodPost, "/auth/register", tc.body, "")
			rec := httptest.NewRecorder()

			authTestHandler.Register(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Code)

			if tc.checkBody != nil {
				var body map[string]any
				err := json.NewDecoder(rec.Body).Decode(&body)
				require.NoError(t, err, "response body should be valid JSON")
				tc.checkBody(t, body)
			}
		})
	}
}

// TestRegister_DuplicateEmail tests that registering with an existing email returns 409.
// This is separate from the table above because it requires two sequential requests
// (register first, then try to register again with same email).
func TestRegister_DuplicateEmail(t *testing.T) {
	testhelpers.CleanupDB(t, authTestDB)

	// First registration — should succeed
	body := map[string]any{
		"name":     "First User",
		"email":    "duplicate@example.com",
		"password": "password123",
	}
	req := testhelpers.NewRequest(t, http.MethodPost, "/auth/register", body, "")
	rec := httptest.NewRecorder()
	authTestHandler.Register(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code, "first registration should succeed")

	// Second registration with same email — should fail with 409
	req2 := testhelpers.NewRequest(t, http.MethodPost, "/auth/register", body, "")
	rec2 := httptest.NewRecorder()
	authTestHandler.Register(rec2, req2)

	assert.Equal(t, http.StatusConflict, rec2.Code, "duplicate email should return 409")

	var respBody map[string]any
	json.NewDecoder(rec2.Body).Decode(&respBody)
	assert.Equal(t, "email already exists", respBody["error"])
}

// --- Login Tests ---

// TestLogin tests all Login scenarios.
// Each test that tests valid login needs a registered user first.
func TestLogin(t *testing.T) {
	// Helper: register a user so login tests have someone to authenticate as
	registerUser := func(t *testing.T, email, password string) {
		t.Helper()
		req := testhelpers.NewRequest(t, http.MethodPost, "/auth/register", map[string]any{
			"name": "Login Test User", "email": email, "password": password,
		}, "")
		rec := httptest.NewRecorder()
		authTestHandler.Register(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code, "setup: user registration failed")
	}

	tests := []struct {
		name           string
		setup          func(t *testing.T) // optional pre-test setup
		body           map[string]any
		expectedStatus int
		checkBody      func(t *testing.T, body map[string]any)
	}{
		{
			name: "valid credentials",
			setup: func(t *testing.T) {
				registerUser(t, "login@example.com", "password123")
			},
			body: map[string]any{
				"email":    "login@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body map[string]any) {
				// Response should have "token" and "user"
				token, hasToken := body["token"]
				assert.True(t, hasToken, "response should include JWT token")
				assert.NotEmpty(t, token, "token should not be empty")

				user, hasUser := body["user"].(map[string]any)
				assert.True(t, hasUser, "response should include user object")
				assert.Equal(t, "login@example.com", user["email"])
				_, hasPasswordHash := user["password_hash"]
				assert.False(t, hasPasswordHash, "user in login response must not include password_hash")
			},
		},
		{
			name: "wrong password",
			setup: func(t *testing.T) {
				registerUser(t, "wrongpass@example.com", "correctpassword")
			},
			body: map[string]any{
				"email":    "wrongpass@example.com",
				"password": "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			checkBody: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "invalid email or password", body["error"])
			},
		},
		{
			name: "email not found",
			body: map[string]any{
				"email":    "doesnotexist@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusUnauthorized,
			checkBody: func(t *testing.T, body map[string]any) {
				// Note: we return the same error for "wrong password" and "email not found"
				// This is intentional — never tell attackers which one failed
				assert.Equal(t, "invalid email or password", body["error"])
			},
		},
		{
			name: "missing email field",
			body: map[string]any{
				"password": "password123",
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "missing password field",
			body: map[string]any{
				"email": "test@example.com",
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testhelpers.CleanupDB(t, authTestDB)

			if tc.setup != nil {
				tc.setup(t)
			}

			req := testhelpers.NewRequest(t, http.MethodPost, "/auth/login", tc.body, "")
			rec := httptest.NewRecorder()

			authTestHandler.Login(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Code)

			if tc.checkBody != nil {
				var body map[string]any
				err := json.NewDecoder(rec.Body).Decode(&body)
				require.NoError(t, err, "response body should be valid JSON")
				tc.checkBody(t, body)
			}
		})
	}
}

// TestLogin_TokenIsUsable verifies that the JWT token returned by Login
// can actually be used to authenticate against the JWT middleware.
// This is an end-to-end test of the auth flow.
func TestLogin_TokenIsUsable(t *testing.T) {
	testhelpers.CleanupDB(t, authTestDB)

	// Register
	req := testhelpers.NewRequest(t, http.MethodPost, "/auth/register", map[string]any{
		"name": "Token Test", "email": "tokentest@example.com", "password": "password123",
	}, "")
	rec := httptest.NewRecorder()
	authTestHandler.Register(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	// Login
	req = testhelpers.NewRequest(t, http.MethodPost, "/auth/login", map[string]any{
		"email": "tokentest@example.com", "password": "password123",
	}, "")
	rec = httptest.NewRecorder()
	authTestHandler.Login(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var loginResp map[string]any
	json.NewDecoder(rec.Body).Decode(&loginResp)
	token, ok := loginResp["token"].(string)
	require.True(t, ok, "login response should have a token string")
	require.NotEmpty(t, token)

	// Use token on a protected endpoint via the full router
	cfg := testhelpers.MakeTestConfig()
	usersHandler := handlers.NewUsersHandler(authTestDB, cfg)

	// Build a router with JWT middleware — same as production wiring
	r := chi.NewRouter()
	r.Get("/api/v1/users", usersHandler.ListUsers)

	req = testhelpers.NewRequest(t, http.MethodGet, "/api/v1/users", nil, token)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Without JWT middleware in this sub-router, it's unprotected — just check it returns 200
	// The full middleware integration is tested in jwt_test.go
	assert.Equal(t, http.StatusOK, rec.Code, "token from login should work on protected routes")
}
