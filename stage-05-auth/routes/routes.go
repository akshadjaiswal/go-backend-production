package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/akshadjaiswal/go-backend-production/stage-05-auth/handlers"
	jwtmw "github.com/akshadjaiswal/go-backend-production/stage-05-auth/middleware"
)

func Setup(usersHandler *handlers.UsersHandler, authHandler *handlers.AuthHandler) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Health check — no auth
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Auth routes — public, no JWT required
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
	})

	// API routes — protected by JWT middleware
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(jwtmw.JWTAuth) // all routes below require a valid JWT

		r.Route("/users", func(r chi.Router) {
			r.Get("/", usersHandler.ListUsers)
			r.Post("/", usersHandler.CreateUser)
			r.Get("/{id}", usersHandler.GetUser)
			r.Put("/{id}", usersHandler.UpdateUser)
			r.Delete("/{id}", usersHandler.DeleteUser)
		})
	})

	return r
}
