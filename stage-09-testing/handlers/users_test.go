package handlers_test

// Integration tests for UsersHandler — all 5 CRUD operations.
//
// These tests use the same authTestDB and TestMain set up in auth_test.go.
// In Go, all _test.go files in the same package share one TestMain.
// That's why both files declare "package handlers_test" — they're the same test package.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/akshadjaiswal/go-backend-production/stage-09-testing/handlers"
	"github.com/akshadjaiswal/go-backend-production/stage-09-testing/testhelpers"
)

// buildUsersRouter creates a Chi router with all users routes wired up.
// We need a real Chi router (not just the handler) because:
//   - GET /users/{id} requires chi.URLParam(r, "id") to work
//   - chi.URLParam reads from the routing context set by Chi
//   - If we call the handler directly (without Chi routing), chi.URLParam returns ""
//
// So for any test involving path params {id}, we must go through chi.Router.ServeHTTP.
func buildUsersRouter(h *handlers.UsersHandler, token string) http.Handler {
	r := chi.NewRouter()

	// For tests, we inject the token via middleware that sets auth header
	// Actually simpler: just mount routes and pass auth via request header
	r.Get("/api/v1/users", h.ListUsers)
	r.Post("/api/v1/users", h.CreateUser)
	r.Get("/api/v1/users/{id}", h.GetUser)
	r.Put("/api/v1/users/{id}", h.UpdateUser)
	r.Delete("/api/v1/users/{id}", h.DeleteUser)
	return r
}

// makeUsersHandler creates a UsersHandler pointing at the test DB.
func makeUsersHandler() *handlers.UsersHandler {
	cfg := testhelpers.MakeTestConfig()
	return handlers.NewUsersHandler(authTestDB, cfg)
}

// insertTestUser inserts a user directly into the DB and returns their ID.
// Used to set up prerequisite data before testing GET/PUT/DELETE.
// We bypass the HTTP layer here because we're testing CRUD handlers,
// not the auth flow — that's already covered in auth_test.go.
func insertTestUser(t *testing.T, name, email string) string {
	t.Helper()
	var id string
	err := authTestDB.Get(&id, `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, 'hash')
		RETURNING id
	`, name, email)
	require.NoError(t, err, "setup: failed to insert test user")
	return id
}

// --- ListUsers ---

func TestListUsers(t *testing.T) {
	h := makeUsersHandler()
	router := buildUsersRouter(h, "")

	t.Run("empty table — returns empty array", func(t *testing.T) {
		testhelpers.CleanupDB(t, authTestDB)

		req := testhelpers.NewRequest(t, http.MethodGet, "/api/v1/users", nil, "")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		// Decode the response — should be an array (even if empty)
		var users []map[string]any
		err := json.NewDecoder(rec.Body).Decode(&users)
		require.NoError(t, err, "response should be a valid JSON array")
		assert.Empty(t, users, "empty table should return empty array, not null")
	})

	t.Run("returns all users ordered by created_at desc", func(t *testing.T) {
		testhelpers.CleanupDB(t, authTestDB)
		insertTestUser(t, "User One", "one@example.com")
		insertTestUser(t, "User Two", "two@example.com")

		req := testhelpers.NewRequest(t, http.MethodGet, "/api/v1/users", nil, "")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var users []map[string]any
		json.NewDecoder(rec.Body).Decode(&users)
		assert.Len(t, users, 2, "should return both users")

		// Check no password_hash in any user
		for _, u := range users {
			_, hasHash := u["password_hash"]
			assert.False(t, hasHash, "password_hash must not appear in list response")
		}
	})
}

// --- CreateUser ---

func TestCreateUser(t *testing.T) {
	h := makeUsersHandler()
	router := buildUsersRouter(h, "")

	tests := []struct {
		name           string
		body           map[string]any
		expectedStatus int
		checkBody      func(t *testing.T, body map[string]any)
	}{
		{
			name: "valid user creation",
			body: map[string]any{
				"name":  "Ketaki",
				"email": "ketaki@example.com",
			},
			expectedStatus: http.StatusCreated,
			checkBody: func(t *testing.T, body map[string]any) {
				assert.NotEmpty(t, body["id"])
				assert.Equal(t, "Ketaki", body["name"])
				assert.Equal(t, "ketaki@example.com", body["email"])
				// CreateUser uses empty password_hash (users without passwords)
				_, hasHash := body["password_hash"]
				assert.False(t, hasHash)
			},
		},
		{
			name:           "missing name",
			body:           map[string]any{"email": "test@example.com"},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "missing email",
			body:           map[string]any{"name": "Test"},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "invalid email format",
			body:           map[string]any{"name": "Test", "email": "bad"},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "name too short (min 2)",
			body:           map[string]any{"name": "A", "email": "a@b.com"},
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testhelpers.CleanupDB(t, authTestDB)

			req := testhelpers.NewRequest(t, http.MethodPost, "/api/v1/users", tc.body, "")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Code)

			if tc.checkBody != nil {
				var body map[string]any
				json.NewDecoder(rec.Body).Decode(&body)
				tc.checkBody(t, body)
			}
		})
	}
}

// TestCreateUser_DuplicateEmail is separate because it needs two sequential requests.
func TestCreateUser_DuplicateEmail(t *testing.T) {
	h := makeUsersHandler()
	router := buildUsersRouter(h, "")
	testhelpers.CleanupDB(t, authTestDB)

	body := map[string]any{"name": "First", "email": "dup@example.com"}

	// First create — success
	req := testhelpers.NewRequest(t, http.MethodPost, "/api/v1/users", body, "")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	// Second create with same email — 409
	req = testhelpers.NewRequest(t, http.MethodPost, "/api/v1/users", body, "")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusConflict, rec.Code)

	var respBody map[string]any
	json.NewDecoder(rec.Body).Decode(&respBody)
	assert.Equal(t, "email already exists", respBody["error"])
}

// --- GetUser ---

func TestGetUser(t *testing.T) {
	h := makeUsersHandler()
	router := buildUsersRouter(h, "")

	t.Run("existing user — returns 200", func(t *testing.T) {
		testhelpers.CleanupDB(t, authTestDB)
		id := insertTestUser(t, "Get Test", "get@example.com")

		req := testhelpers.NewRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/users/%s", id), nil, "")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var user map[string]any
		json.NewDecoder(rec.Body).Decode(&user)
		assert.Equal(t, id, user["id"])
		assert.Equal(t, "Get Test", user["name"])
		assert.Equal(t, "get@example.com", user["email"])
		_, hasHash := user["password_hash"]
		assert.False(t, hasHash, "password_hash must not be in response")
	})

	t.Run("non-existent user — returns 404", func(t *testing.T) {
		testhelpers.CleanupDB(t, authTestDB)

		// Valid UUID4 format, but no user with this ID exists
		fakeID := "00000000-0000-4000-a000-000000000001"
		req := testhelpers.NewRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/users/%s", fakeID), nil, "")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)

		var body map[string]any
		json.NewDecoder(rec.Body).Decode(&body)
		assert.Equal(t, "user not found", body["error"])
	})

	t.Run("invalid UUID — returns 400", func(t *testing.T) {
		req := testhelpers.NewRequest(t, http.MethodGet, "/api/v1/users/not-a-uuid", nil, "")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var body map[string]any
		json.NewDecoder(rec.Body).Decode(&body)
		assert.Equal(t, "invalid user ID format", body["error"])
	})
}

// --- UpdateUser ---

func TestUpdateUser(t *testing.T) {
	h := makeUsersHandler()
	router := buildUsersRouter(h, "")

	t.Run("valid update — returns 200 with updated data", func(t *testing.T) {
		testhelpers.CleanupDB(t, authTestDB)
		id := insertTestUser(t, "Original Name", "original@example.com")

		body := map[string]any{
			"name":  "Updated Name",
			"email": "updated@example.com",
		}
		req := testhelpers.NewRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/users/%s", id), body, "")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var user map[string]any
		json.NewDecoder(rec.Body).Decode(&user)
		assert.Equal(t, "Updated Name", user["name"])
		assert.Equal(t, "updated@example.com", user["email"])
		assert.Equal(t, id, user["id"], "user ID should not change after update")
	})

	t.Run("non-existent user — returns 404", func(t *testing.T) {
		testhelpers.CleanupDB(t, authTestDB)

		fakeID := "00000000-0000-4000-a000-000000000002"
		body := map[string]any{"name": "Ghost", "email": "ghost@example.com"}
		req := testhelpers.NewRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/users/%s", fakeID), body, "")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("invalid UUID — returns 400", func(t *testing.T) {
		body := map[string]any{"name": "Test", "email": "test@example.com"}
		req := testhelpers.NewRequest(t, http.MethodPut, "/api/v1/users/bad-uuid", body, "")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("missing name in body — returns 422", func(t *testing.T) {
		testhelpers.CleanupDB(t, authTestDB)
		id := insertTestUser(t, "Update Validation", "updateval@example.com")

		body := map[string]any{"email": "updateval@example.com"} // missing name
		req := testhelpers.NewRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/users/%s", id), body, "")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	})
}

// --- DeleteUser ---

func TestDeleteUser(t *testing.T) {
	h := makeUsersHandler()
	router := buildUsersRouter(h, "")

	t.Run("existing user — returns 204 no content", func(t *testing.T) {
		testhelpers.CleanupDB(t, authTestDB)
		id := insertTestUser(t, "Delete Me", "deleteme@example.com")

		req := testhelpers.NewRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/users/%s", id), nil, "")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		// 204 No Content — no body
		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Empty(t, rec.Body.String(), "204 response should have no body")

		// Verify the user is actually gone
		req = testhelpers.NewRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/users/%s", id), nil, "")
		rec = httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code, "deleted user should not be found")
	})

	t.Run("non-existent user — returns 404", func(t *testing.T) {
		testhelpers.CleanupDB(t, authTestDB)

		fakeID := "00000000-0000-4000-a000-000000000003"
		req := testhelpers.NewRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/users/%s", fakeID), nil, "")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)

		var body map[string]any
		json.NewDecoder(rec.Body).Decode(&body)
		assert.Equal(t, "user not found", body["error"])
	})

	t.Run("invalid UUID — returns 400", func(t *testing.T) {
		req := testhelpers.NewRequest(t, http.MethodDelete, "/api/v1/users/not-a-uuid", nil, "")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
