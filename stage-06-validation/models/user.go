package models

import "time"

// User maps to the users table — same as stage 05.
type User struct {
	ID           string    `json:"id"         db:"id"`
	Name         string    `json:"name"       db:"name"`
	Email        string    `json:"email"      db:"email"`
	PasswordHash string    `json:"-"          db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// RegisterRequest — validate tags define the rules.
// The validator library reads these tags just like json/db tags.
//
// Rules used:
//   required   — field must be present and non-empty
//   min=N      — minimum string length N
//   max=N      — maximum string length N
//   email      — must be a valid email format (has @ and domain)
type RegisterRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	// max=72 for password because bcrypt silently truncates at 72 bytes
}

// LoginRequest — both fields required, email must be valid format.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse — returned after successful login.
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// CreateUserRequest — used when creating a user directly (not via register).
type CreateUserRequest struct {
	Name  string `json:"name"  validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email"`
}

// UpdateUserRequest — all fields optional on update (omitempty would be used
// for partial updates, but we keep it simple here with full replacement).
type UpdateUserRequest struct {
	Name  string `json:"name"  validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email"`
}
