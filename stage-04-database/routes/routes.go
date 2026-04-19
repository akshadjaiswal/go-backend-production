package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	custommiddleware "github.com/akshadjaiswal/go-backend-production/stage-04-database/middleware"
	"github.com/akshadjaiswal/go-backend-production/stage-04-database/handlers"
)

func Setup(usersHandler *handlers.UsersHandler) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(custommiddleware.AuthGuard)

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
