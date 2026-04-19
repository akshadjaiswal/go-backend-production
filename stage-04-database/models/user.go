package models

import "time"

// User maps to the users table in PostgreSQL.
// Notice two sets of tags:
//   `json:"id"`  — controls JSON field names (for API responses)
//   `db:"id"`    — controls DB column mapping (for sqlx scanning)
//
// Both tags on the same field is very common in Go backends.
type User struct {
	ID        string    `json:"id"         db:"id"`
	Name      string    `json:"name"       db:"name"`
	Email     string    `json:"email"      db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateUserRequest is what the client sends when creating a user.
// We separate input types from the full model — the client shouldn't
// send id, created_at, updated_at (the DB generates those).
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UpdateUserRequest is what the client sends when updating a user.
type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
