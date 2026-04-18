package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/akshadjaiswal/go-backend-production/stage-03-middleware/models"
)

var store = map[string]models.User{
	"1": {ID: "1", Name: "Akshad Jaiswal", Email: "akshad@example.com"},
	"2": {ID: "2", Name: "Sid Tiwatne", Email: "sid@example.com"},
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func ListUsers(w http.ResponseWriter, r *http.Request) {
	users := make([]models.User, 0, len(store))
	for _, u := range store {
		users = append(users, u)
	}
	writeJSON(w, http.StatusOK, users)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if user.Name == "" || user.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and email are required"})
		return
	}
	user.ID = fmt.Sprintf("%d", len(store)+1)
	store[user.ID] = user
	writeJSON(w, http.StatusCreated, user)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, ok := store[id]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, ok := store[id]; !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	var updated models.User
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	updated.ID = id
	store[id] = updated
	writeJSON(w, http.StatusOK, updated)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, ok := store[id]; !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	delete(store, id)
	w.WriteHeader(http.StatusNoContent)
}
