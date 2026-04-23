// main.go is the entry point for the stage-10 Go backend.
//
// It wires together config → logger → database → handlers → routes → server.
// Each dependency is created once and passed to the next — this is called
// dependency injection. No global variables, no hidden state.
//
// In Docker, this binary is built by the Dockerfile and run by the container.
// The container gets its config from docker-compose.yml's `environment:` block.
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/akshadjaiswal/go-backend-production/stage-10-deployment/config"
	appdb "github.com/akshadjaiswal/go-backend-production/stage-10-deployment/db"
	"github.com/akshadjaiswal/go-backend-production/stage-10-deployment/handlers"
	applogger "github.com/akshadjaiswal/go-backend-production/stage-10-deployment/logger"
	"github.com/akshadjaiswal/go-backend-production/stage-10-deployment/routes"
)

func main() {
	// 1. Load config — reads from .env (local dev) or env vars (Docker).
	// Fail fast if required vars are missing — better than a cryptic runtime error.
	cfg, err := config.Load()
	if err != nil {
		// Logger not set up yet — use raw stderr for this one message.
		os.Stderr.WriteString("Config error: " + err.Error() + "\n")
		os.Stderr.WriteString("In Docker: check the `environment:` block in docker-compose.yml\n")
		os.Stderr.WriteString("Locally: copy .env.example to .env and fill in values\n")
		os.Exit(1)
	}

	// 2. Set up structured logger.
	// ENV=dev        → text format (readable in terminal)
	// ENV=production → JSON format (docker logs, log aggregators)
	applogger.Setup(cfg.Env)

	slog.Info("starting server",
		slog.String("env", cfg.Env),
		slog.String("port", cfg.Port),
		slog.Int("jwt_expiry_hours", cfg.JWTExpiryHours),
	)

	// 3. Connect to database.
	// In Docker: DATABASE_URL host is "postgres" (service name resolved by Docker DNS).
	// Locally: DATABASE_URL host is "localhost".
	database, err := appdb.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database",
			slog.String("error", err.Error()),
			slog.String("hint", "in Docker: postgres container may still be starting — depends_on should prevent this"),
		)
		os.Exit(1)
	}
	defer database.Close()

	slog.Info("connected to database")

	// 4. Create handlers — inject DB and config.
	usersHandler := handlers.NewUsersHandler(database, cfg)
	authHandler := handlers.NewAuthHandler(database, cfg)

	// 5. Wire routes — returns a chi.Router (implements http.Handler).
	r := routes.Setup(usersHandler, authHandler, cfg)

	slog.Info("server ready",
		slog.String("address", "http://localhost:"+cfg.Port),
		slog.String("health", "http://localhost:"+cfg.Port+"/health"),
	)

	// 6. Start the HTTP server — blocks until the process is killed.
	// In Docker, SIGTERM from `docker compose down` will reach this process
	// (because CMD uses exec form, not shell form — see Dockerfile).
	if err := http.ListenAndServe(cfg.ServerAddress(), r); err != nil {
		slog.Error("server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
