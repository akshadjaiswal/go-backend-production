package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/akshadjaiswal/go-backend-production/stage-09-testing/config"
	appdb "github.com/akshadjaiswal/go-backend-production/stage-09-testing/db"
	"github.com/akshadjaiswal/go-backend-production/stage-09-testing/handlers"
	applogger "github.com/akshadjaiswal/go-backend-production/stage-09-testing/logger"
	"github.com/akshadjaiswal/go-backend-production/stage-09-testing/routes"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		// Logger not set up yet — use plain stderr for this one error
		os.Stderr.WriteString("Config error: " + err.Error() + "\n")
		os.Stderr.WriteString("See .env.example for required variables.\n")
		os.Exit(1)
	}

	// Set up structured logger FIRST — before any other logs.
	// Dev → text format (readable in terminal)
	// Production → JSON format (parseable by log tools)
	applogger.Setup(cfg.Env)

	slog.Info("starting server",
		slog.String("env", cfg.Env),
		slog.String("port", cfg.Port),
		slog.Int("jwt_expiry_hours", cfg.JWTExpiryHours),
	)

	database, err := appdb.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer database.Close()

	slog.Info("connected to database")

	usersHandler := handlers.NewUsersHandler(database, cfg)
	authHandler := handlers.NewAuthHandler(database, cfg)

	r := routes.Setup(usersHandler, authHandler, cfg)

	slog.Info("server ready", slog.String("address", "http://localhost:"+cfg.Port))

	if err := http.ListenAndServe(cfg.ServerAddress(), r); err != nil {
		slog.Error("server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
