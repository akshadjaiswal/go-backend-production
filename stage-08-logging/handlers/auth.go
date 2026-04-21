package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"

	"github.com/akshadjaiswal/go-backend-production/stage-08-logging/config"
	"github.com/akshadjaiswal/go-backend-production/stage-08-logging/models"
	v "github.com/akshadjaiswal/go-backend-production/stage-08-logging/validator"
)

type AuthHandler struct {
	DB  *sqlx.DB
	cfg *config.Config // injected — no more hardcoded secrets
}

func NewAuthHandler(db *sqlx.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{DB: db, cfg: cfg}
}

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
	writeJSON(w, http.StatusCreated, user)
}

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

	var user models.User
	err := h.DB.Get(&user, `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM users WHERE email = $1
	`, req.Email)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

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

// generateToken uses cfg.JWTSecret and cfg.JWTExpiryHours — both from config
func (h *AuthHandler) generateToken(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"name":    user.Name,
		"exp":     time.Now().Add(time.Duration(h.cfg.JWTExpiryHours) * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWTSecret))
}
