package main

import (
	"fmt"
	"net/http"
	"os"

	appdb "github.com/akshadjaiswal/go-backend-production/stage-04-database/db"
	"github.com/akshadjaiswal/go-backend-production/stage-04-database/handlers"
	"github.com/akshadjaiswal/go-backend-production/stage-04-database/routes"
)

func main() {
	// Read the DSN from an environment variable.
	// os.Getenv returns "" if the variable isn't set.
	// Stage 07 (Config) will handle this more cleanly with a proper config package.
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Fallback default for local development
		dsn = "postgres://postgres:postgres@localhost:5432/go_backend_production?sslmode=disable"
	}

	// Connect to PostgreSQL
	database, err := appdb.Connect(dsn)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1) // os.Exit(1) = exit with error code (non-zero = failure)
	}
	defer database.Close() // close the connection pool when main() returns

	// Create the handler with the DB injected
	usersHandler := handlers.NewUsersHandler(database)

	// Setup routes, passing in our handler
	r := routes.Setup(usersHandler)

	fmt.Println("Stage 04 — Server starting on http://localhost:8080")
	fmt.Println("Database: PostgreSQL via sqlx")

	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}
