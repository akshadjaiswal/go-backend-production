package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"github.com/akshadjaiswal/go-backend-production/stage-05-auth/models"
)

type UsersHandler struct {
	DB *sqlx.DB
}

func NewUsersHandler(db *sqlx.DB) *UsersHandler {
	return &UsersHandler{DB: db}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *UsersHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	err := h.DB.Select(&users, `SELECT id, name, email, created_at, updated_at FROM users ORDER BY created_at DESC`)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch users"})
		return
	}
	writeJSON(w, http.StatusOK, users)
}

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
	err := h.DB.Get(&user, `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, '')
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

func (h *UsersHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var user models.User
	err := h.DB.Get(&user, `SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1`, id)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *UsersHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	var user models.User
	err := h.DB.Get(&user, `
		UPDATE users SET name = $1, email = $2, updated_at = NOW()
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

func (h *UsersHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	result, err := h.DB.Exec(`DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete user"})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

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
