// package routes wires all HTTP routes and middleware together.
//
// Route structure:
//   GET  /health             → public health check (used by Docker healthcheck)
//   POST /auth/register      → public
//   POST /auth/login         → public
//   GET  /api/v1/users       → JWT-protected
//   POST /api/v1/users       → JWT-protected
//   GET  /api/v1/users/{id}  → JWT-protected
//   PUT  /api/v1/users/{id}  → JWT-protected
//   DEL  /api/v1/users/{id}  → JWT-protected
package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/akshadjaiswal/go-backend-production/stage-10-deployment/config"
	"github.com/akshadjaiswal/go-backend-production/stage-10-deployment/handlers"
	"github.com/akshadjaiswal/go-backend-production/stage-10-deployment/middleware"
)

// Setup builds the chi router with all routes and middleware registered.
// Returns chi.Router which implements http.Handler — passed directly to http.ListenAndServe.
func Setup(usersHandler *handlers.UsersHandler, authHandler *handlers.AuthHandler, cfg *config.Config) chi.Router {
	r := chi.NewRouter()

	// Global middleware — runs for EVERY request
	r.Use(middleware.RequestLogger) // structured JSON logging
	r.Use(chimw.Recoverer)          // catches panics, returns 500 instead of crashing

	// Public health check — no auth, no DB.
	// Docker's HEALTHCHECK and load balancers hit this.
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"env":    cfg.Env, // "production" in Docker, "dev" locally
		})
	})

	// Auth routes — public, no JWT required
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
	})

	// Protected API routes — JWT middleware applied to the whole group
	r.Route("/api/v1", func(r chi.Router) {
		// r.Use() here only applies to routes inside this Route block
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
