# Stage 07 — Config & Environment Variables

> **Goal:** Replace every hardcoded value (DB URL, JWT secret, port) with environment variables. Use a `.env` file for local development. Fail fast at startup if required config is missing.

---

## What changed from Stage 06?

| Stage 06 | Stage 07 |
|----------|----------|
| `var jwtSecret = []byte("hardcoded...")` | `cfg.JWTSecret` from `.env` |
| `dsn = "postgres://...@localhost..."` in main | `DATABASE_URL` from `.env` |
| Port hardcoded as `":8080"` | `PORT` from `.env` |
| No fail-fast on missing config | `os.Exit(1)` with clear message if required vars missing |
| Secrets could leak into git | `.env` is gitignored, `.env.example` is committed |

---

## Project structure

```
stage-07-config/
├── .env                     ← your local secrets (gitignored, never commit)
├── .env.example             ← template committed to git (no real values)
├── main.go                  ← loads config first, passes to everything
├── config/
│   └── config.go            ← Load(), Config struct, validation
├── handlers/
│   ├── auth.go              ← takes *config.Config, uses cfg.JWTSecret
│   └── users.go             ← takes *config.Config
├── middleware/
│   └── jwt.go               ← NewJWTAuth(cfg) — struct-based middleware
└── ...
```

---

## Key concepts

### 1. `.env` file

A plain text file with `KEY=VALUE` pairs:

```env
DATABASE_URL=postgres://akshad@localhost:5432/mydb?sslmode=disable
JWT_SECRET=some-long-random-string
PORT=8080
```

**Rules:**
- `.env` → gitignored (contains real secrets)
- `.env.example` → committed (shows what vars are needed, no real values)
- Real env vars always override `.env` values

### 2. `godotenv.Load()`

```go
_ = godotenv.Load() // loads .env into process environment
```

- Reads `.env` and calls `os.Setenv()` for each key
- If `.env` doesn't exist → no error (fine in production)
- Existing env vars are NOT overwritten — real env always wins
- The `_` ignores the error intentionally

### 3. Config struct — single source of truth

```go
type Config struct {
    Port           string
    Env            string
    DatabaseURL    string
    JWTSecret      string
    JWTExpiryHours int
}
```

One struct. Loaded once. Passed everywhere. No scattered `os.Getenv()` calls across the codebase.

### 4. Fail fast on missing required config

```go
cfg.DatabaseURL = os.Getenv("DATABASE_URL")
if cfg.DatabaseURL == "" {
    return nil, fmt.Errorf("DATABASE_URL is required but not set")
}
```

If you start the server without `DATABASE_URL`, you get:
```
Config error: DATABASE_URL is required but not set
Make sure you have a .env file or the required environment variables set.
See .env.example for required variables.
```

This is much better than starting up and crashing on the first DB query.

### 5. Injecting config into handlers

```go
// Before (stage 06) — hardcoded secret in handler
var jwtSecret = []byte("jwt-secret-key-change-in-production")

// After (stage 07) — config injected via constructor
type AuthHandler struct {
    DB  *sqlx.DB
    cfg *config.Config
}

func NewAuthHandler(db *sqlx.DB, cfg *config.Config) *AuthHandler {
    return &AuthHandler{DB: db, cfg: cfg}
}

// Used inside handler:
token.SignedString([]byte(h.cfg.JWTSecret))
```

### 6. Struct-based middleware

```go
type JWTMiddleware struct {
    cfg *config.Config
}

func NewJWTAuth(cfg *config.Config) *JWTMiddleware {
    return &JWTMiddleware{cfg: cfg}
}

// Handler is the actual middleware func
func (m *JWTMiddleware) Handler(next http.Handler) http.Handler { ... }

// In routes:
r.Use(jwtmw.NewJWTAuth(cfg).Handler)
```

Middleware needs config too — using a struct lets us inject it cleanly.

### 7. `.gitignore` for `.env`

The repo `.gitignore` already has:
```
.env
.env.*
!.env.example
```

- `.env` → ignored
- `.env.dev`, `.env.prod` → ignored
- `.env.example` → NOT ignored (committed as template)

---

## Setup

### 1. Copy .env.example to .env
```bash
cp .env.example .env
```

### 2. Edit .env with your values
```bash
# Update DATABASE_URL with your username
DATABASE_URL=postgres://$(whoami)@localhost:5432/go_backend_production_stage07?sslmode=disable
JWT_SECRET=pick-any-long-random-string
```

### 3. Create DB and run migration
```bash
createdb go_backend_production_stage07
psql -d go_backend_production_stage07 -f migrations/001_create_users.sql
```

### 4. Start server
```bash
cd stage-07-config
go run main.go
```

You'll see:
```
Connected to PostgreSQL
Stage 07 — Server starting on http://localhost:8080
Environment: dev
JWT expiry: 24 hours
```

### 5. Override with env vars directly
```bash
PORT=9090 go run main.go
# → Server starting on http://localhost:9090
```

### 6. Test missing config (should fail fast)
```bash
# Temporarily rename .env and unset var
mv .env .env.bak
go run main.go
# → Config error: DATABASE_URL is required but not set
mv .env.bak .env
```

---

## What's missing (coming next)

| Missing | Added in |
|---------|----------|
| All logs are plain text prints | Stage 08 — Structured JSON Logging |
| No request correlation in logs | Stage 08 — Logging |
