// package config loads all application configuration from environment variables.
// It reads a .env file first (for local dev), then actual env vars override those.
// The whole app gets one *Config — no more scattered os.Getenv() calls everywhere.
package config

import (
	"fmt"
	"os"
	"strconv"

	// godotenv loads a .env file into the process environment.
	// It does NOT override existing env vars — so real env vars always win.
	"github.com/joho/godotenv"
)

// Config holds every configurable value in the application.
// All hardcoded secrets from previous stages now live here.
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

// Load reads the .env file (if present) and environment variables,
// validates required fields, applies defaults, and returns a *Config.
//
// Call this once at the start of main() — fail fast if config is invalid.
func Load() (*Config, error) {
	// godotenv.Load() reads .env and sets each key as an env var.
	// If .env doesn't exist, it just does nothing — not an error.
	// This means in production you don't need a .env file at all;
	// you just set env vars directly (Docker, Kubernetes, etc.)
	_ = godotenv.Load()

	cfg := &Config{}

	// --- Required fields ---
	// If these are missing, we cannot start. Better to crash now than
	// fail on the first request with a confusing error.

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required but not set")
	}

	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required but not set")
	}

	// --- Optional fields with defaults ---

	cfg.Port = os.Getenv("PORT")
	if cfg.Port == "" {
		cfg.Port = "8080" // sensible default
	}

	cfg.Env = os.Getenv("ENV")
	if cfg.Env == "" {
		cfg.Env = "dev"
	}

	// strconv.Atoi converts a string to int — returns error if not a valid number
	jwtExpiryStr := os.Getenv("JWT_EXPIRY_HOURS")
	if jwtExpiryStr == "" {
		cfg.JWTExpiryHours = 24 // default: 24 hours
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

// ServerAddress returns the full address string for http.ListenAndServe.
// e.g. ":8080"
func (c *Config) ServerAddress() string {
	return ":" + c.Port
}
