// package handlers contains our HTTP handler functions.
// Each handler is responsible for one specific action (list, create, get, update, delete).
// Separating handlers from routing keeps each file focused and small.
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	// chi gives us URLParam() to read path parameters like {id}
	"github.com/go-chi/chi/v5"

	// importing our own models package using the module path from go.mod
	"github.com/akshadjaiswal/go-backend-production/stage-02-routing/models"
)

// store is our in-memory "database" — a map of user ID → User.
// map[string]models.User means: keys are strings, values are User structs.
// In Stage 04 we'll replace this with a real PostgreSQL database.
var store = map[string]models.User{
	"1": {ID: "1", Name: "Akshad Jaiswal", Email: "akshad@example.com"},
	"2": {ID: "2", Name: "Sid Tiwatne", Email: "sid@example.com"},
}

// writeJSON is a small helper to avoid repeating header + encode in every handler.
// Notice it's lowercase writeJSON — private to this package (not exported).
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// ListUsers handles GET /api/v1/users
// Returns all users as a JSON array.
func ListUsers(w http.ResponseWriter, r *http.Request) {
	// Convert map values to a slice so JSON encodes it as an array, not an object.
	users := make([]models.User, 0, len(store))
	for _, u := range store {
		users = append(users, u)
	}
	writeJSON(w, http.StatusOK, users)
}

// CreateUser handles POST /api/v1/users
// Reads JSON body, creates a new user, stores it, returns 201 Created.
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User

	// json.NewDecoder(r.Body).Decode(&user) reads the request body and
	// populates our user struct. &user means "pointer to user" — Decode needs
	// a pointer so it can modify the actual variable (like passing by reference).
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	// Basic validation — name and email are required
	if user.Name == "" || user.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "name and email are required",
		})
		return
	}

	// Simple ID generation using current store length + 1.
	// In production we'd use UUIDs (Stage 06 covers this).
	user.ID = fmt.Sprintf("%d", len(store)+1)
	store[user.ID] = user

	// 201 Created — standard response when a new resource is created
	writeJSON(w, http.StatusCreated, user)
}

// GetUser handles GET /api/v1/users/{id}
// chi.URLParam reads the {id} path parameter from the URL.
func GetUser(w http.ResponseWriter, r *http.Request) {
	// chi.URLParam(r, "id") extracts the {id} from the URL path.
	// e.g. GET /api/v1/users/42 → id = "42"
	id := chi.URLParam(r, "id")

	user, ok := store[id]
	// ok is false if the key doesn't exist in the map — Go's safe map lookup
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "user not found",
		})
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// UpdateUser handles PUT /api/v1/users/{id}
// Reads body, updates the user in store, returns updated user.
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	_, ok := store[id]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "user not found",
		})
		return
	}

	var updated models.User
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	// Keep the original ID — don't let the client change it via the body
	updated.ID = id
	store[id] = updated

	writeJSON(w, http.StatusOK, updated)
}

// DeleteUser handles DELETE /api/v1/users/{id}
// Removes the user from store, returns 204 No Content (no body).
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	_, ok := store[id]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "user not found",
		})
		return
	}

	// delete() is Go's built-in function to remove a key from a map
	delete(store, id)

	// 204 No Content — standard for successful DELETE with no response body
	w.WriteHeader(http.StatusNoContent)
}
