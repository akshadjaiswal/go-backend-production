package main

import (
	"fmt"
	"net/http"
	"os"

	appdb "github.com/akshadjaiswal/go-backend-production/stage-06-validation/db"
	"github.com/akshadjaiswal/go-backend-production/stage-06-validation/handlers"
	"github.com/akshadjaiswal/go-backend-production/stage-06-validation/routes"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/go_backend_production_stage06?sslmode=disable"
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

	fmt.Println("Stage 06 — Server starting on http://localhost:8080")
	fmt.Println("Validation: go-playground/validator with readable error messages")

	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}
