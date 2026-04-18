# Stage 03 — Middleware

> **Goal:** Write custom middleware from scratch. Understand how middleware chains work, how to pass data through context, and how to protect routes with an auth guard.

---

## What changed from Stage 02?

| Stage 02 | Stage 03 |
|----------|----------|
| Used Chi's built-in `middleware.Logger` | Custom Logger that captures status code |
| No request tracing | RequestID middleware — every request gets a unique ID |
| No CORS headers | CORS middleware — browser requests work |
| No auth | AuthGuard — `/api/v1/*` requires `X-API-Key` header |

---

## Project structure

```
stage-03-middleware/
├── main.go
├── routes/
│   └── routes.go          ← global middleware + route-level AuthGuard
├── handlers/
│   └── users.go           ← same as stage 02
├── middleware/
│   ├── request_id.go      ← generates unique ID per request
│   ├── logger.go          ← logs method, path, status, duration
│   ├── cors.go            ← CORS headers for browser requests
│   └── auth.go            ← API key guard
├── models/
│   └── user.go
├── requests.http
└── README.md
```

---

## The middleware signature

Every middleware in Go looks like this:

```go
func MyMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // do something BEFORE the handler
        next.ServeHTTP(w, r) // call the next middleware/handler
        // do something AFTER the handler
    })
}
```

- Takes `next http.Handler` — the next thing in the chain
- Returns a new `http.Handler` that wraps it
- Call `next.ServeHTTP(w, r)` to pass the request forward
- Don't call `next` to stop the chain (e.g. in AuthGuard when key is invalid)

---

## Key concepts

### 1. Middleware chain & order

```
Request → RequestID → Logger → CORS → AuthGuard → Handler → Response
```

Order matters:
- **RequestID first** — so Logger can read the ID
- **Logger before AuthGuard** — so even rejected requests get logged
- **AuthGuard last** — so it can block before reaching the handler

In `routes.go`:
```go
r.Use(middleware.RequestID)  // 1st
r.Use(middleware.Logger)     // 2nd
r.Use(middleware.CORS)       // 3rd
// AuthGuard only inside /api/v1 group
r.Use(middleware.AuthGuard)  // 4th (route-level)
```

### 2. context — passing data through the chain

```go
// Attach data to the request context
ctx := context.WithValue(r.Context(), middleware.RequestIDKey, "abc-123")
next.ServeHTTP(w, r.WithContext(ctx))

// Read it anywhere downstream (handler, other middleware)
id := middleware.GetRequestID(r.Context())
```

Think of context like a bag that travels with the request. You can put things in it and read them anywhere downstream.

### 3. Wrapping ResponseWriter to capture status code

The standard `http.ResponseWriter` doesn't let you read the status code after it's set. Our Logger needs it for logging. Solution — wrap it:

```go
type responseWriter struct {
    http.ResponseWriter       // embed original — inherit all methods
    statusCode int            // we capture the code here
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code      // save it
    rw.ResponseWriter.WriteHeader(code) // still write it
}
```

### 4. Global vs route-level middleware

```go
r := chi.NewRouter()

// Global — runs on ALL routes including /health
r.Use(middleware.RequestID)
r.Use(middleware.Logger)
r.Use(middleware.CORS)

r.Get("/health", ...) // ← Logger runs, AuthGuard does NOT

r.Route("/api/v1", func(r chi.Router) {
    r.Use(middleware.AuthGuard) // ← only inside this group
    r.Route("/users", ...)      // ← AuthGuard runs here
})
```

### 5. Stopping the chain early (AuthGuard)

```go
if apiKey != validAPIKey {
    w.WriteHeader(http.StatusUnauthorized)
    json.NewEncoder(w).Encode(...)
    return // ← do NOT call next.ServeHTTP — chain stops here
}
next.ServeHTTP(w, r) // ← only reached if key is valid
```

---

## How to run

```bash
cd stage-03-middleware
go run main.go
```

You'll see logs like:
```
[2026-04-18 10:00:01] GET /health → 200 (45µs) | req_id=1713430801-3421
[2026-04-18 10:00:02] GET /api/v1/users → 401 (12µs) | req_id=1713430802-8823
[2026-04-18 10:00:03] GET /api/v1/users → 200 (89µs) | req_id=1713430803-1122
```

---

## Test the endpoints

Open `requests.http` in VS Code and try each request. Key ones to understand:

```bash
# Health — no auth needed
curl http://localhost:8080/health

# No API key → 401
curl http://localhost:8080/api/v1/users

# Wrong API key → 401
curl http://localhost:8080/api/v1/users -H "X-API-Key: wrong"

# Valid API key → 200
curl http://localhost:8080/api/v1/users -H "X-API-Key: secret-key-123"
```

Notice the `X-Request-ID` header in every response — that's our RequestID middleware.

---

## What's missing (coming next)

| Missing | Added in |
|---------|----------|
| Real database | Stage 04 — PostgreSQL |
| JWT tokens instead of API keys | Stage 05 — Auth |
| API key from environment variable | Stage 07 — Config |
