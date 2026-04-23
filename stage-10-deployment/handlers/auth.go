// Auth handlers: register and login.
//
// Register: hash password with bcrypt, insert user, return user JSON.
// Login: look up user by email, compare password, issue JWT, return token + user.
//
// JWT secret comes from config (injected) — never hardcoded.
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"

	"github.com/akshadjaiswal/go-backend-production/stage-10-deployment/config"
	"github.com/akshadjaiswal/go-backend-production/stage-10-deployment/models"
	v "github.com/akshadjaiswal/go-backend-production/stage-10-deployment/validator"
)

// AuthHandler holds DB and config for auth-related handlers.
type AuthHandler struct {
	DB  *sqlx.DB
	cfg *config.Config
}

// NewAuthHandler is the constructor — used in main() to wire everything together.
func NewAuthHandler(db *sqlx.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{DB: db, cfg: cfg}
}

// Register handles POST /auth/register
// Creates a new user account with a hashed password.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
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

	// bcrypt hashes the password with a random salt.
	// DefaultCost = 10 rounds — secure enough, ~100ms on modern hardware.
	// Never store plain-text passwords. Never use MD5/SHA for passwords.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
		return
	}

	var user models.User
	err = h.DB.Get(&user, `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, created_at, updated_at
	`, req.Name, req.Email, string(hashedPassword))
	if err != nil {
		if containsUniqueViolation(err) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "email already exists"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
		return
	}

	// Return 201 Created with the user (no password hash — json:"-" hides it)
	writeJSON(w, http.StatusCreated, user)
}

// Login handles POST /auth/login
// Validates credentials and returns a JWT on success.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
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

	// Look up the user — we need password_hash for comparison.
	// Note: we fetch password_hash here (not in ListUsers/GetUser) because we need it.
	var user models.User
	err := h.DB.Get(&user, `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM users WHERE email = $1
	`, req.Email)
	if err == sql.ErrNoRows {
		// Don't say "user not found" — that reveals whether an email is registered.
		// Always say "invalid email or password" for security.
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	// bcrypt.CompareHashAndPassword handles timing-safe comparison.
	// Returns non-nil error if passwords don't match.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
		return
	}

	token, err := h.generateToken(user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		return
	}

	writeJSON(w, http.StatusOK, models.LoginResponse{Token: token, User: user})
}

// generateToken creates a signed JWT for the given user.
// Claims include user_id and email (for use in downstream handlers).
// Expiry comes from config — configurable without code changes.
func (h *AuthHandler) generateToken(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"name":    user.Name,
		// exp = expiry time as Unix timestamp — jwt library checks this automatically
		"exp": time.Now().Add(time.Duration(h.cfg.JWTExpiryHours) * time.Hour).Unix(),
		// iat = issued at — useful for debugging token age
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Sign with the secret — any change to the payload or secret invalidates the signature
	return token.SignedString([]byte(h.cfg.JWTSecret))
}
