package models

import "time"

// User maps to the users table.
// PasswordHash has json:"-" — the dash means NEVER include this in JSON output.
// You never want to send password hashes to clients, even accidentally.
type User struct {
	ID           string    `json:"id"         db:"id"`
	Name         string    `json:"name"       db:"name"`
	Email        string    `json:"email"      db:"email"`
	PasswordHash string    `json:"-"          db:"password_hash"` // never serialized to JSON
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// RegisterRequest is what the client sends to /auth/register
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest is what the client sends to /auth/login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse is what we return after a successful login
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// CreateUserRequest and UpdateUserRequest — same as stage 04
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
