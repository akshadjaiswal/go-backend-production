// package handlers contains HTTP handler functions for the API.
//
// Handler struct pattern: handlers are methods on a struct that holds dependencies.
// This avoids global variables and makes testing easy (inject a test DB).
//
// This file: CRUD operations for users (admin API, JWT-protected).
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"github.com/akshadjaiswal/go-backend-production/stage-10-deployment/config"
	"github.com/akshadjaiswal/go-backend-production/stage-10-deployment/models"
	v "github.com/akshadjaiswal/go-backend-production/stage-10-deployment/validator"
)

// UsersHandler holds dependencies for user-related handlers.
// Created once in main() and reused for every request.
type UsersHandler struct {
	DB  *sqlx.DB
	cfg *config.Config
}

// NewUsersHandler is the constructor — used in main() to wire everything together.
func NewUsersHandler(db *sqlx.DB, cfg *config.Config) *UsersHandler {
	return &UsersHandler{DB: db, cfg: cfg}
}

// writeJSON writes a JSON response with the given status code.
// Private to this package — each handler package defines its own copy.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// ListUsers handles GET /api/v1/users
// Returns all users sorted by creation date (newest first).
func (h *UsersHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	// We select specific columns (not *) to avoid returning password_hash.
	// Even though json:"-" hides it, it's better practice to not fetch it at all.
	if err := h.DB.Select(&users, `SELECT id, name, email, created_at, updated_at FROM users ORDER BY created_at DESC`); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch users"})
		return
	}
	writeJSON(w, http.StatusOK, users)
}

// CreateUser handles POST /api/v1/users
// Creates a user without a password (admin operation — not self-registration).
func (h *UsersHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Validate all fields against the struct tags before touching the DB.
	if err := v.Validate.Struct(req); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error": "validation failed", "fields": v.FormatErrors(err),
		})
		return
	}

	var user models.User
	// RETURNING * gets the full row back in one query — no separate SELECT needed.
	err := h.DB.Get(&user, `
		INSERT INTO users (name, email, password_hash) VALUES ($1, $2, '')
		RETURNING id, name, email, created_at, updated_at
	`, req.Name, req.Email)
	if err != nil {
		if containsUniqueViolation(err) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "email already exists"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

// GetUser handles GET /api/v1/users/{id}
func (h *UsersHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Validate UUID format before querying — prevents garbage hitting the DB.
	if err := v.Validate.Var(id, "required,uuid4"); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user ID format"})
		return
	}

	var user models.User
	err := h.DB.Get(&user, `SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1`, id)
	if err == sql.ErrNoRows {
		// sql.ErrNoRows is the specific "not found" signal — check this before generic error.
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// UpdateUser handles PUT /api/v1/users/{id}
func (h *UsersHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := v.Validate.Var(id, "required,uuid4"); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user ID format"})
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if err := v.Validate.Struct(req); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error": "validation failed", "fields": v.FormatErrors(err),
		})
		return
	}

	var user models.User
	err := h.DB.Get(&user, `
		UPDATE users SET name = $1, email = $2, updated_at = NOW()
		WHERE id = $3 RETURNING id, name, email, created_at, updated_at
	`, req.Name, req.Email, id)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update user"})
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// DeleteUser handles DELETE /api/v1/users/{id}
// Returns 204 No Content on success (no body — nothing to return after deletion).
func (h *UsersHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := v.Validate.Var(id, "required,uuid4"); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user ID format"})
		return
	}

	result, err := h.DB.Exec(`DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete user"})
		return
	}

	// RowsAffected tells us if the DELETE actually hit a row.
	// If 0, the user didn't exist — return 404 instead of silently succeeding.
	rows, _ := result.RowsAffected()
	if rows == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 — success, no body
}

// containsUniqueViolation checks if a DB error is a unique constraint violation.
// PostgreSQL error code 23505 = unique_violation.
// We check the error string because the pq driver wraps it in a generic error.
func containsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	for _, phrase := range []string{"23505", "unique constraint", "duplicate key"} {
		for i := 0; i <= len(msg)-len(phrase); i++ {
			if msg[i:i+len(phrase)] == phrase {
				return true
			}
		}
	}
	return false
}
