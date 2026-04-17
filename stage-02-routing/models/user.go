// package models holds our data types (structs).
// Separating models from handlers keeps code organized —
// any package can import models without importing handler logic.
package models

// User represents a user in our system.
// This is our data shape — like a TypeScript interface.
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
