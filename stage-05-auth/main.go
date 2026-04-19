package main

import (
	"fmt"
	"net/http"
	"os"

	appdb "github.com/akshadjaiswal/go-backend-production/stage-05-auth/db"
	"github.com/akshadjaiswal/go-backend-production/stage-05-auth/handlers"
	"github.com/akshadjaiswal/go-backend-production/stage-05-auth/routes"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/go_backend_production_stage05?sslmode=disable"
	}

	database, err := appdb.Connect(dsn)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	usersHandler := handlers.NewUsersHandler(database)
	authHandler := handlers.NewAuthHandler(database)

	r := routes.Setup(usersHandler, authHandler)

	fmt.Println("Stage 05 — Server starting on http://localhost:8080")
	fmt.Println("Auth: JWT (Bearer token)")
	fmt.Println("")
	fmt.Println("Public routes:")
	fmt.Println("  POST /auth/register")
	fmt.Println("  POST /auth/login")
	fmt.Println("")
	fmt.Println("Protected routes (requires Authorization: Bearer <token>):")
	fmt.Println("  GET    /api/v1/users")
	fmt.Println("  POST   /api/v1/users")
	fmt.Println("  GET    /api/v1/users/{id}")
	fmt.Println("  PUT    /api/v1/users/{id}")
	fmt.Println("  DELETE /api/v1/users/{id}")

	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}
