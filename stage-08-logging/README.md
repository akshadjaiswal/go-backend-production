# Stage 08 — Structured JSON Logging

> **Goal:** Replace all plain `fmt.Println` and Chi's default logger with Go's built-in `slog` package. Every log line becomes a structured JSON object with consistent fields — ready for production log tools like Datadog, CloudWatch, or Grafana Loki.

---

## What changed from Stage 07?

| Stage 07 | Stage 08 |
|----------|----------|
| `fmt.Println("Connected to PostgreSQL")` | `slog.Info("connected to database")` |
| Chi's plain text request logger | Custom structured JSON request logger |
| No log levels | `DEBUG`, `INFO`, `WARN`, `ERROR` |
| Unstructured strings | Key-value fields: `method`, `path`, `status`, `duration` |
| Same format in dev and prod | Text in dev, JSON in production |
| No request tracing in logs | `request_id` on every log line |

---

## What structured logs look like

**Dev (text format — easy to read in terminal):**
```
time=2026-04-21T10:00:01Z level=INFO msg="starting server" env=dev port=8080
time=2026-04-21T10:00:02Z level=INFO msg=request method=GET path=/health status=200 duration=165µs request_id=430802-3421
time=2026-04-21T10:00:03Z level=WARN msg=request method=GET path=/api/v1/users status=401 duration=12µs request_id=430803-8823
```

**Production (JSON format — machine-parseable):**
```json
{"time":"2026-04-21T10:00:01Z","level":"INFO","msg":"starting server","env":"dev","port":"8080"}
{"time":"2026-04-21T10:00:02Z","level":"INFO","msg":"request","method":"GET","path":"/health","status":200,"duration":"165µs","request_id":"430802-3421"}
{"time":"2026-04-21T10:00:03Z","level":"WARN","msg":"request","method":"GET","path":"/api/v1/users","status":401,"duration":"12µs","request_id":"430803-8823"}
```

---

## Project structure

```
stage-08-logging/
├── main.go                  ← calls logger.Setup() first, uses slog throughout
├── logger/
│   └── logger.go            ← NEW: sets up global slog, dev vs prod format
├── middleware/
│   ├── jwt.go               ← same as stage-07
│   └── logger.go            ← NEW: structured request logger middleware
└── ... (everything else same as stage-07)
```

Only two new files. Everything from stage-07 carries forward.

---

## Key concepts

### 1. `slog` — Go's built-in structured logger

`slog` was added to the Go standard library in Go 1.21. **No external package needed.**

```go
import "log/slog"

slog.Info("user created", slog.String("user_id", "abc-123"), slog.String("email", "x@x.com"))
slog.Error("db failed", slog.String("error", err.Error()))
slog.Debug("cache hit", slog.String("key", "users:list"))
slog.Warn("rate limit close", slog.Int("remaining", 5))
```

Each call produces one structured log line with the message + all key-value fields attached.

### 2. Log levels

| Level | Use for |
|-------|---------|
| `DEBUG` | Detailed internals — DB queries, cache hits. Dev only. |
| `INFO` | Normal operations — server started, request completed, user created |
| `WARN` | Something unexpected but handled — 4xx errors, retried operations |
| `ERROR` | Something broke — 5xx errors, DB failures, panics |

In `dev` we see all 4 levels. In `production` we only see INFO, WARN, ERROR — DEBUG is filtered out to reduce noise.

### 3. Text vs JSON handler

```go
// Dev — human readable
handler = slog.NewTextHandler(os.Stdout, opts)
// Output: time=2026-04-21T10:00:01Z level=INFO msg="request" status=200

// Production — machine readable
handler = slog.NewJSONHandler(os.Stdout, opts)
// Output: {"time":"2026-04-21T10:00:01Z","level":"INFO","msg":"request","status":200}
```

`slog.SetDefault(slog.New(handler))` sets the global logger — all `slog.Info()` calls anywhere in the app use it automatically.

### 4. Structured fields vs plain strings

```go
// Before — unstructured, impossible to filter in production
fmt.Printf("request %s %s → %d in %s\n", method, path, status, duration)

// After — structured, filterable by any field
slog.Info("request",
    slog.String("method", method),
    slog.String("path", path),
    slog.Int("status", status),
    slog.String("duration", duration.String()),
    slog.String("request_id", requestID),
)
```

In Datadog or CloudWatch you can now filter: `status>=500` or `path="/api/v1/users"` or `level=ERROR` — completely impossible with plain strings.

### 5. Request logger middleware

Our `middleware.RequestLogger` replaces Chi's plain text `middleware.Logger`:

```
Request arrives
  → RequestLogger: start timer, generate request_id
  → Handler runs (auth, DB query, response written)
  → RequestLogger: log method + path + status + duration + request_id
```

Log level is chosen automatically based on status code:
```go
status >= 500 → slog.Error   // something broke on our side
status >= 400 → slog.Warn    // client did something wrong
else          → slog.Info    // normal successful request
```

This means in production you can set up alerts on `level=ERROR` to catch 5xx automatically — no manual threshold configuration needed.

### 6. `request_id` for tracing

Every request gets a unique ID in its log line and response header:
```
X-Request-ID: 430802-3421
```

If a user reports a bug, they can give you their `X-Request-ID`. You grep your logs for that ID and see exactly what happened for that specific request — even with thousands of concurrent users in the logs.

### 7. Why `slog.String()` and not just string arguments?

```go
// Both work, but the typed form is preferred:
slog.Info("user created", "user_id", "abc")          // works
slog.Info("user created", slog.String("user_id", "abc")) // preferred
```

The typed form (`slog.String`, `slog.Int`, `slog.Bool`, etc.) is type-safe and avoids subtle bugs where you pass an odd number of key-value pairs.

---

## Setup

### 1. Copy .env and update
```bash
cp .env.example .env
# Edit DATABASE_URL — replace "youruser" with your Mac username (run: whoami)
```

### 2. Create DB and run migration
```bash
createdb go_backend_production_stage08
psql -d go_backend_production_stage08 -f migrations/001_create_users.sql
```

### 3. Start server (dev mode — text logs)
```bash
cd stage-08-logging
go run main.go
```

You'll see:
```
time=... level=INFO msg="starting server" env=dev port=8080 jwt_expiry_hours=24
time=... level=INFO msg="connected to database"
time=... level=INFO msg="server ready" address=http://localhost:8080
```

### 4. Test production JSON logs
```bash
ENV=production go run main.go
```
Every log line is now a JSON object.

---

## Testing — what to watch for

Open a terminal with the server running and watch the logs as you test:

| Action | Expected log level | Why |
|--------|-------------------|-----|
| `GET /health` | INFO | Normal 200 response |
| `GET /api/v1/users` (no token) | WARN | 401 — client error |
| `GET /api/v1/users/bad-uuid` | WARN | 400 — client error |
| `POST /auth/register` (valid) | INFO | Normal 201 response |
| `GET /api/v1/users` (valid token) | INFO | Normal 200 response |

Each log line also includes `request_id` — notice it matches the `X-Request-ID` response header.

---

## What's missing (coming next)

| Missing | Added in |
|---------|----------|
| No automated tests for any of our code | Stage 09 — Testing |
| Docker / containerization | Stage 10 — Deployment |
