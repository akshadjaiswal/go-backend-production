# Stage 10 — Deployment with Docker

## Goal

Learn how to package a Go backend into a Docker container and run the full stack (app + database) with a single command using docker-compose.

By the end of this stage you understand:
- What Docker is and why it matters
- Multi-stage builds — how to get a ~15MB production image from a ~700MB Go toolchain
- docker-compose — wiring multiple containers together
- How Docker networking works (service-name DNS)
- How to auto-apply DB migrations on first container start
- How the same Go code works locally (`.env` file) and in Docker (injected env vars)

---

## What changed from stage-09

| Area | Stage 09 | Stage 10 |
|------|----------|----------|
| Run command | `go run ./stage-09-testing/` | `docker compose up --build` |
| Database setup | manual `createdb` + `psql -f migration.sql` | automatic on first `docker compose up` |
| Config source | `.env` file | `environment:` block in `docker-compose.yml` |
| Port binding | `localhost:8080` directly | `localhost:8080` → container:8080 |
| Test files | yes (unit + integration) | none (covered in stage-09) |
| New files | none | `Dockerfile`, `docker-compose.yml`, `.dockerignore` |

The Go source code is identical to stage-09. The only changes are:
1. Import paths: `stage-09-testing` → `stage-10-deployment`
2. Three new Docker files

---

## Project structure

```
stage-10-deployment/
├── main.go                    ← entry point, wires everything
├── config/config.go           ← reads env vars (same works locally + Docker)
├── db/db.go                   ← PostgreSQL connection pool
├── logger/logger.go           ← structured slog (JSON in production)
├── models/user.go             ← data shapes
├── handlers/
│   ├── users.go               ← CRUD user handlers
│   └── auth.go                ← register + login handlers
├── middleware/
│   ├── jwt.go                 ← JWT auth middleware
│   └── logger.go              ← structured request logging
├── validator/validator.go     ← input validation wrapper
├── routes/routes.go           ← all route wiring
├── migrations/
│   └── 001_create_users.sql   ← schema (auto-applied by Docker postgres)
├── Dockerfile                 ← multi-stage build
├── docker-compose.yml         ← app + postgres wired together
├── .dockerignore              ← keeps build context lean
├── .env.example               ← template for local dev (without Docker)
└── requests.http              ← VS Code REST Client test file
```

---

## Key concepts

### What is Docker?

Docker packages your app and its dependencies into a **container** — a lightweight, isolated process that runs the same way on any machine.

Without Docker:
- "Works on my machine" — different OS, different Go version, different PostgreSQL
- Manual setup: install Go, install Postgres, create DB, run migrations...

With Docker:
- One command: `docker compose up --build`
- Identical environment on every machine
- App and database both containerised — no local installs needed

### What is a container?

A container is like a very lightweight virtual machine. It runs an isolated process with its own filesystem, network, and environment — but shares the host OS kernel (unlike a full VM).

```
Your Mac
├── Docker daemon
│   ├── Container: postgres  (runs PostgreSQL 16, port 5432 internal)
│   └── Container: app       (runs ./app binary, port 8080 → host 8080)
```

### Multi-stage builds

```dockerfile
# Stage 1: Builder — has Go toolchain (~700MB)
FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download        # cached layer — only re-runs if go.mod changes
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app ./stage-10-deployment/

# Stage 2: Runner — just Alpine Linux (~5MB) + the binary
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /app .  # copy ONLY the compiled binary
CMD ["./app"]
```

**Why two stages?**

| Stage | Image | Size |
|-------|-------|------|
| Builder | `golang:1.23-alpine` | ~700MB (Go toolchain, stdlib, headers) |
| Runner | `alpine:3.19` + binary | ~15MB total |

The final image only contains the compiled binary. Go compiles to a single static binary — no runtime, no interpreter, no external libraries needed.

**CGO_ENABLED=0** disables C bindings, producing a fully static binary. Required on Alpine (which uses musl libc, not glibc). Without this, the binary would crash with "not found" errors inside the container.

**GOOS=linux** cross-compiles for Linux even if you're building on macOS or Windows.

### Docker layer caching

```dockerfile
COPY go.mod go.sum ./   # layer 1 — cached until go.mod changes
RUN go mod download     # layer 2 — only re-runs when layer 1 changes
COPY . .                # layer 3 — changes every time you edit code
RUN go build ...        # layer 4 — only re-runs when layer 3 changes
```

If you only change a `.go` file, layers 1 and 2 hit the cache. Only layers 3 and 4 re-run. This makes rebuilds fast.

### docker-compose networking

```yaml
services:
  postgres:
    image: postgres:16-alpine
    # ...

  app:
    environment:
      DATABASE_URL: postgres://goapp:secret@postgres:5432/...
                                             ^^^^^^^^
                                     service name = hostname
```

Docker Compose creates a private network for all services in the file. Each service can reach any other by its **service name** — Docker's internal DNS resolves it.

The `app` container connects to `postgres:5432` (not `localhost:5432`). `localhost` inside a container refers to that container itself, not the host or other containers.

### Health checks and depends_on

```yaml
postgres:
  healthcheck:
    test: ["CMD-SHELL", "pg_isready -U goapp -d go_backend_production_stage10"]
    interval: 5s
    retries: 5

app:
  depends_on:
    postgres:
      condition: service_healthy  # app waits for this
```

Without `depends_on: condition: service_healthy`, the app container starts immediately — but PostgreSQL takes a few seconds to initialise. The app would fail to connect and crash.

With the health check, Compose waits until `pg_isready` succeeds (PostgreSQL is actually accepting connections) before starting the app.

### Auto-migrations via initdb.d

```yaml
postgres:
  volumes:
    - ./migrations:/docker-entrypoint-initdb.d
```

The official `postgres` Docker image runs all `.sql` and `.sh` files in `/docker-entrypoint-initdb.d/` when it initialises a **fresh** data volume.

This means:
- First `docker compose up` → schema is created automatically
- No manual `createdb` or `psql -f migration.sql` needed
- Subsequent `docker compose up` → initdb.d is skipped (data volume already exists)
- `docker compose down -v` → removes volume → next `up` re-runs migrations

### Same config code, two environments

```go
// config/config.go
_ = godotenv.Load()  // silently fails if .env doesn't exist
cfg.DatabaseURL = os.Getenv("DATABASE_URL")
```

**Local dev:** `godotenv.Load()` reads `.env` file, sets env vars, then `os.Getenv()` reads them.

**Docker:** no `.env` file inside the container. Docker injects vars from `docker-compose.yml` `environment:` block directly into the process. `godotenv.Load()` silently fails — `os.Getenv()` still works.

Zero code changes between environments. The config is just environment variables either way.

### JSON logging in production

```yaml
# docker-compose.yml
environment:
  ENV: production  # triggers JSON log format
```

```go
// logger/logger.go
if env == "production" {
    handler = slog.NewJSONHandler(os.Stdout, opts)
}
```

In Docker, logs go to `stdout` and are captured by Docker's log driver. When you run `docker compose logs`, you see them. In a real deployment, a log aggregator (Datadog, CloudWatch, Loki) picks up the JSON lines and indexes every field.

Example production log line:
```json
{"time":"2026-04-23T10:00:01Z","level":"INFO","msg":"request","method":"GET","path":"/health","status":200,"duration":"245µs","request_id":"430802-3421","remote_addr":"172.18.0.1:52341"}
```

---

## Setup

### Prerequisites

- Docker Desktop (includes both `docker` and `docker compose`) — [install](https://www.docker.com/products/docker-desktop/)
- VS Code with REST Client extension (for `requests.http`)

No local Go or PostgreSQL needed — Docker handles everything.

### Run the full stack

```bash
cd stage-10-deployment

# Build the Go binary and start both containers
docker compose up --build
```

First run takes ~30–60 seconds:
1. Downloads `golang:1.23-alpine` and `postgres:16-alpine` images
2. Builds the Go binary (multi-stage)
3. Starts PostgreSQL, runs migrations
4. Starts the app (waits for postgres health check)

Subsequent runs are much faster — Docker caches layers.

### Verify it's working

```bash
# In a second terminal
curl http://localhost:8080/health
# → {"env":"production","status":"ok"}
```

### View logs

```bash
# All containers
docker compose logs -f

# Just the app
docker compose logs -f app

# Just postgres
docker compose logs -f postgres
```

### Stop the stack

```bash
# Stop containers, keep DB data
docker compose down

# Stop containers AND wipe DB (fresh start next time)
docker compose down -v
```

### Rebuild after code changes

```bash
docker compose up --build
```

The `--build` flag forces a rebuild of the app image. Without it, Docker uses the cached image from last time.

---

## Test flow

After `docker compose up --build`, use `requests.http` in VS Code.

### Step 1 — Health check
Send `GET /health`. Expect `{"status":"ok","env":"production"}`.

### Step 2 — Register
Send `POST /auth/register` with name, email, password. Expect `201` with user JSON.

### Step 3 — Login
Send `POST /auth/login`. Expect `200` with `{"token":"...", "user":{...}}`.
Copy the token value.

### Step 4 — Set token
In `requests.http`, replace `PASTE_TOKEN_HERE` with your copied token.

### Step 5 — Test protected routes
- `GET /api/v1/users` → list all users
- `POST /api/v1/users` → create user (admin, no password)
- Get a user ID from the list, replace `PASTE_USER_ID_HERE`
- `GET /api/v1/users/{id}` → get one user
- `PUT /api/v1/users/{id}` → update user
- `DELETE /api/v1/users/{id}` → delete user

### Error cases
- Register duplicate email → 409
- Invalid UUID → 400
- No auth header → 401
- Invalid token → 401

---

## Useful Docker commands

```bash
# List running containers
docker ps

# List all images
docker images

# Remove unused images (free disk space)
docker image prune

# Shell into the app container (for debugging)
docker compose exec app sh

# Shell into the postgres container
docker compose exec postgres psql -U goapp -d go_backend_production_stage10

# Rebuild only the app image (not postgres)
docker compose build app
docker compose up app
```

---

## What's missing

| What | Why it matters | When to add |
|------|---------------|-------------|
| HEALTHCHECK in Dockerfile | Docker can restart unhealthy containers automatically | Production deployments |
| Graceful shutdown | In-flight requests finish before server stops | Before any real traffic |
| Secrets management | JWT secret in docker-compose.yml is not ideal | Use Docker secrets or a vault |
| Reverse proxy (nginx/Caddy) | TLS termination, rate limiting, static files | Before exposing to internet |
| CI/CD | Automated build + push to registry on git push | Before team deployment |
| Kubernetes | Scaling, rolling deploys, self-healing | When you outgrow single-machine Docker |
| Non-root user in container | Security best practice | Any real production image |
