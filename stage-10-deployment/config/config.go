// package config loads all application configuration from environment variables.
//
// Two environments, same code:
//   - Local dev:  reads a .env file (via godotenv), then env vars override
//   - Docker:     no .env file inside the container — Docker injects env vars directly
//
// godotenv.Load() silently does nothing if .env doesn't exist, so the same
// config code works in both cases without any changes. This is intentional.
package config

import (
	"fmt"
	"os"
	"strconv"

	// godotenv reads a .env file and sets each key as an env var.
	// It does NOT override vars that are already set — real env vars always win.
	// In Docker, vars come from docker-compose.yml's `environment:` block.
	"github.com/joho/godotenv"
)

// Config holds every configurable value in the application.
// Loaded once at startup in main() and passed to everything via constructors.
// Never call os.Getenv() anywhere else — always use Config fields.
type Config struct {
	// Server
	Port string // e.g. "8080"
	Env  string // "dev" or "production"

	// Database
	DatabaseURL string // full PostgreSQL DSN

	// JWT
	JWTSecret      string // secret key for signing/verifying tokens
	JWTExpiryHours int    // how many hours until a token expires
}

// Load reads config from environment (and optionally a .env file),
// validates required fields, applies defaults, and returns *Config.
//
// Fail fast: if any required field is missing, return an error immediately.
// Better to crash at startup than fail on the first request.
func Load() (*Config, error) {
	// Silently ignore error if .env doesn't exist.
	// In Docker, there's no .env — vars come from the container environment.
	// In local dev, .env provides the vars.
	_ = godotenv.Load()

	cfg := &Config{}

	// --- Required fields (app cannot start without these) ---

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required but not set")
	}

	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required but not set")
	}

	// --- Optional fields with sensible defaults ---

	cfg.Port = os.Getenv("PORT")
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	cfg.Env = os.Getenv("ENV")
	if cfg.Env == "" {
		cfg.Env = "dev"
	}

	jwtExpiryStr := os.Getenv("JWT_EXPIRY_HOURS")
	if jwtExpiryStr == "" {
		cfg.JWTExpiryHours = 24
	} else {
		hours, err := strconv.Atoi(jwtExpiryStr)
		if err != nil {
			return nil, fmt.Errorf("JWT_EXPIRY_HOURS must be a number, got: %s", jwtExpiryStr)
		}
		cfg.JWTExpiryHours = hours
	}

	return cfg, nil
}

// IsDev returns true when running in development mode.
func (c *Config) IsDev() bool {
	return c.Env == "dev"
}

// ServerAddress returns the address string for http.ListenAndServe.
// e.g. ":8080"
func (c *Config) ServerAddress() string {
	return ":" + c.Port
}
