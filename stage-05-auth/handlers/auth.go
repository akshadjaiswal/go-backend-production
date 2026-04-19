package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"

	"github.com/akshadjaiswal/go-backend-production/stage-05-auth/models"
)

// jwtSecret is the key used to sign and verify JWT tokens.
// Stage 07 (Config) will load this from an environment variable.
// NEVER hardcode this in production.
var jwtSecret = []byte("jwt-secret-key-change-in-production")

// AuthHandler holds the DB connection for auth operations.
type AuthHandler struct {
	DB *sqlx.DB
}

func NewAuthHandler(db *sqlx.DB) *AuthHandler {
	return &AuthHandler{DB: db}
}

// Register handles POST /auth/register
// Creates a new user with a bcrypt-hashed password.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name, email and password are required"})
		return
	}

	if len(req.Password) < 6 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 6 characters"})
		return
	}

	// bcrypt.GenerateFromPassword hashes the password.
	// bcrypt.DefaultCost = 10 — the "cost factor", higher = slower = more secure.
	// bcrypt is intentionally slow to make brute-force attacks hard.
	// It also automatically handles salting (no need to generate a salt yourself).
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

	// Return the user but NOT a token — user must login separately.
	// This is the standard pattern: register ≠ auto-login.
	writeJSON(w, http.StatusCreated, user)
}

// Login handles POST /auth/login
// Verifies credentials and returns a signed JWT token.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		return
	}

	// Fetch user including password_hash (note: no json output because of json:"-" tag)
	var user models.User
	err := h.DB.Get(&user, `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM users WHERE email = $1
	`, req.Email)

	if err == sql.ErrNoRows {
		// Use the same error message for "user not found" and "wrong password"
		// to prevent email enumeration attacks (attacker can't tell which is wrong)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	// bcrypt.CompareHashAndPassword compares the plain password against the stored hash.
	// Returns nil if they match, error if not.
	// This is safe against timing attacks — always takes the same amount of time.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
		return
	}

	// Generate JWT token
	token, err := generateToken(user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		return
	}

	writeJSON(w, http.StatusOK, models.LoginResponse{
		Token: token,
		User:  user,
	})
}

// generateToken creates a signed JWT token for the given user.
func generateToken(user models.User) (string, error) {
	// jwt.MapClaims is the payload of the token — the data embedded inside.
	// Anyone can decode the payload (it's just base64) — don't put sensitive data here.
	// The signature ensures it wasn't tampered with.
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"name":    user.Name,
		// exp = expiry time — Unix timestamp. Token invalid after this.
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		// iat = issued at — when the token was created
		"iat": time.Now().Unix(),
	}

	// jwt.NewWithClaims creates an unsigned token with HS256 algorithm.
	// HS256 = HMAC-SHA256 — symmetric signing (same key to sign and verify).
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// SignedString signs the token with our secret key and returns the final string.
	// Format: base64(header).base64(payload).base64(signature)
	return token.SignedString(jwtSecret)
}
