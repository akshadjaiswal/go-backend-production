package models

import "time"

// User represents a row in the users table.
// Tags:
//   json:"..."  — field name in JSON responses
//   db:"..."    — column name sqlx maps to
//   json:"-"    — password_hash never appears in JSON (security!)
type User struct {
	ID           string    `json:"id"         db:"id"`
	Name         string    `json:"name"       db:"name"`
	Email        string    `json:"email"      db:"email"`
	PasswordHash string    `json:"-"          db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// RegisterRequest is the body expected by POST /auth/register.
// validate tags are used by go-playground/validator to enforce rules.
type RegisterRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// LoginRequest is the body expected by POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse is returned on successful login.
// Contains the JWT and the user's public info.
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// CreateUserRequest is used by POST /api/v1/users (admin create, no password).
type CreateUserRequest struct {
	Name  string `json:"name"  validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email"`
}

// UpdateUserRequest is used by PUT /api/v1/users/:id.
type UpdateUserRequest struct {
	Name  string `json:"name"  validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email"`
}
