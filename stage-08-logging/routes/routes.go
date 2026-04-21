package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/akshadjaiswal/go-backend-production/stage-08-logging/config"
	"github.com/akshadjaiswal/go-backend-production/stage-08-logging/handlers"
	"github.com/akshadjaiswal/go-backend-production/stage-08-logging/middleware"
)

func Setup(usersHandler *handlers.UsersHandler, authHandler *handlers.AuthHandler, cfg *config.Config) chi.Router {
	r := chi.NewRouter()

	// Use our structured JSON logger instead of Chi's plain text logger
	r.Use(middleware.RequestLogger)
	r.Use(chimw.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"env":    cfg.Env,
		})
	})

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.NewJWTAuth(cfg).Handler)

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
