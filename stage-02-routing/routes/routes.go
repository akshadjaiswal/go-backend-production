// package routes defines all the URL routes for the application.
// Keeping routes separate from main.go and handlers keeps things clean —
// you can see the entire API structure at a glance in one file.
package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/akshadjaiswal/go-backend-production/stage-02-routing/handlers"
)

// Setup creates and returns a configured Chi router.
// main.go calls this and passes the result to http.ListenAndServe.
func Setup() chi.Router {
	// chi.NewRouter() creates a new router instance
	r := chi.NewRouter()

	// middleware.Logger logs every request: method, path, status, duration
	// This is our first taste of middleware — runs before every handler
	r.Use(middleware.Logger)

	// middleware.Recoverer catches panics in handlers and returns 500
	// instead of crashing the entire server — essential for production
	r.Use(middleware.Recoverer)

	// Simple health check at root level (not under /api/v1)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// r.Route creates a route group — all routes inside share the prefix /api/v1
	// This is how you version your API: v1, v2, etc.
	r.Route("/api/v1", func(r chi.Router) {

		// Another nested group for /api/v1/users
		// All user-related routes go here
		r.Route("/users", func(r chi.Router) {
			r.Get("/", handlers.ListUsers)       // GET    /api/v1/users
			r.Post("/", handlers.CreateUser)     // POST   /api/v1/users

			// {id} is a path parameter — Chi captures whatever is in that position
			// and makes it available via chi.URLParam(r, "id")
			r.Get("/{id}", handlers.GetUser)     // GET    /api/v1/users/{id}
			r.Put("/{id}", handlers.UpdateUser)  // PUT    /api/v1/users/{id}
			r.Delete("/{id}", handlers.DeleteUser) // DELETE /api/v1/users/{id}
		})
	})

	return r
}
