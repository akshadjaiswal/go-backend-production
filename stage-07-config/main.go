package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/akshadjaiswal/go-backend-production/stage-07-config/config"
	appdb "github.com/akshadjaiswal/go-backend-production/stage-07-config/db"
	"github.com/akshadjaiswal/go-backend-production/stage-07-config/handlers"
	"github.com/akshadjaiswal/go-backend-production/stage-07-config/routes"
)

func main() {
	// Load config first — fail fast if required values are missing.
	// config.Load() reads .env then environment variables.
	cfg, err := config.Load()
	if err != nil {
		// Print a clear error and exit — don't start a broken server
		fmt.Printf("Config error: %v\n", err)
		fmt.Println("Make sure you have a .env file or the required environment variables set.")
		fmt.Println("See .env.example for required variables.")
		os.Exit(1)
	}

	database, err := appdb.Connect(cfg.DatabaseURL)
	if err != nil {
		fmt.Printf("Database error: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// Pass cfg into handlers — no more hardcoded secrets
	usersHandler := handlers.NewUsersHandler(database, cfg)
	authHandler := handlers.NewAuthHandler(database, cfg)

	r := routes.Setup(usersHandler, authHandler, cfg)

	fmt.Printf("Stage 07 — Server starting on http://localhost:%s\n", cfg.Port)
	fmt.Printf("Environment: %s\n", cfg.Env)
	fmt.Printf("JWT expiry: %d hours\n", cfg.JWTExpiryHours)

	if err := http.ListenAndServe(cfg.ServerAddress(), r); err != nil {
		panic(err)
	}
}
