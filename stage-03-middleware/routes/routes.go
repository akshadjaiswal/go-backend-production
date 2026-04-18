package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/akshadjaiswal/go-backend-production/stage-03-middleware/handlers"
	"github.com/akshadjaiswal/go-backend-production/stage-03-middleware/middleware"
)

func Setup() chi.Router {
	r := chi.NewRouter()

	// Global middleware — runs on EVERY request, including /health
	// Order matters: RequestID first so Logger can read the ID
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.CORS)

	// /health has no auth — anyone can call it
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// /api/v1 routes are protected by AuthGuard
	// r.Group creates a sub-router that inherits the parent's middleware
	// then we add AuthGuard ONLY for this group
	r.Route("/api/v1", func(r chi.Router) {
		// AuthGuard is applied only inside this route group
		// /health above is NOT affected
		r.Use(middleware.AuthGuard)

		r.Route("/users", func(r chi.Router) {
			r.Get("/", handlers.ListUsers)
			r.Post("/", handlers.CreateUser)
			r.Get("/{id}", handlers.GetUser)
			r.Put("/{id}", handlers.UpdateUser)
			r.Delete("/{id}", handlers.DeleteUser)
		})
	})

	return r
}
