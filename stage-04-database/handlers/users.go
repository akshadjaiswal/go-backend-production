package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"github.com/akshadjaiswal/go-backend-production/stage-04-database/models"
)

// UsersHandler holds the DB connection.
// This is the "handler struct" pattern — instead of global variables,
// we attach dependencies (like the DB) to a struct and use methods as handlers.
// Benefits: easier to test, no global state, clear dependencies.
type UsersHandler struct {
	DB *sqlx.DB
}

// NewUsersHandler creates a new UsersHandler with the given DB.
// main.go calls this and passes the result to routes.
func NewUsersHandler(db *sqlx.DB) *UsersHandler {
	return &UsersHandler{DB: db}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// ListUsers handles GET /api/v1/users
// db.Select scans multiple rows directly into a slice of structs.
func (h *UsersHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User

	// db.Select runs the query and scans all rows into the slice.
	// sqlx uses the `db` struct tags to map columns → fields automatically.
	err := h.DB.Select(&users, `SELECT id, name, email, created_at, updated_at FROM users ORDER BY created_at DESC`)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch users"})
		return
	}

	writeJSON(w, http.StatusOK, users)
}

// CreateUser handles POST /api/v1/users
// Inserts a new row and returns the created user (with DB-generated id + timestamps).
func (h *UsersHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Name == "" || req.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and email are required"})
		return
	}

	var user models.User

	// RETURNING * tells PostgreSQL to return the full inserted row.
	// db.Get scans a single row into a struct — perfect for INSERT...RETURNING.
	// $1, $2 are parameterized placeholders — NEVER concatenate user input into SQL.
	err := h.DB.Get(&user, `
		INSERT INTO users (name, email)
		VALUES ($1, $2)
		RETURNING id, name, email, created_at, updated_at
	`, req.Name, req.Email)

	if err != nil {
		// Check for unique constraint violation (duplicate email)
		// pq error code 23505 = unique_violation
		if err.Error() != "" && containsUniqueViolation(err) {
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

	var user models.User

	// db.Get scans exactly one row. Returns sql.ErrNoRows if nothing found.
	err := h.DB.Get(&user, `SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1`, id)
	if err == sql.ErrNoRows {
		// sql.ErrNoRows is the standard "not found" error in Go's database/sql
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

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	var user models.User

	err := h.DB.Get(&user, `
		UPDATE users
		SET name = $1, email = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id, name, email, created_at, updated_at
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
func (h *UsersHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// db.Exec runs a query without scanning results — good for DELETE/UPDATE
	// when you don't need RETURNING.
	result, err := h.DB.Exec(`DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete user"})
		return
	}

	// RowsAffected tells us how many rows were deleted.
	// If 0 — the user didn't exist.
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// containsUniqueViolation checks if a PostgreSQL error is a unique constraint violation.
func containsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	for _, phrase := range []string{"23505", "unique constraint", "duplicate key"} {
		if len(msg) >= len(phrase) {
			for i := 0; i <= len(msg)-len(phrase); i++ {
				if msg[i:i+len(phrase)] == phrase {
					return true
				}
			}
		}
	}
	return false
}
