# Stage 02 — Routing with Chi

> **Goal:** Replace the standard library mux with Chi router. Add path parameters, route grouping, method-based routing, and split code across multiple packages.

---

## What changed from Stage 01?

| Stage 01 | Stage 02 |
|----------|----------|
| Single `main.go` file | Code split into `handlers/`, `routes/`, `models/` |
| `http.NewServeMux()` | Chi router (`go-chi/chi/v5`) |
| No path params | `{id}` path params via `chi.URLParam()` |
| No route grouping | `/api/v1/users` route group |
| No middleware | Chi's built-in Logger + Recoverer middleware |

---

## Why Chi?

Go's standard mux (even in Go 1.22+) doesn't support:
- Named path params like `{id}` in a clean, readable way
- Route grouping / nesting
- Middleware chaining

Chi adds all of this while keeping the **exact same handler signature** as standard library:
```go
func(w http.ResponseWriter, r *http.Request)
```
So everything you learned in Stage 01 still applies.

---

## Project structure

```
stage-02-routing/
├── main.go              ← entry point, starts the server
├── routes/
│   └── routes.go        ← all route definitions in one place
├── handlers/
│   └── users.go         ← one function per endpoint
├── models/
│   └── user.go          ← User struct (data shape)
├── requests.http        ← test all endpoints in VS Code
└── README.md
```

### Why split into packages?

- **models/** — data shapes only. No logic. Any package can import it.
- **handlers/** — HTTP logic only. Reads request, writes response.
- **routes/** — wiring only. Maps URLs to handlers.
- **main.go** — startup only. Creates server, nothing else.

Each file has one job. This is how real Go backends are structured.

---

## Key concepts introduced

### 1. Installing external packages with `go get`
```bash
go get github.com/go-chi/chi/v5
```
This downloads Chi and adds it to `go.mod` and `go.sum`.
- `go.mod` — lists your dependencies (like `package.json`)
- `go.sum` — checksums for security (like `package-lock.json`)

### 2. Path parameters
```go
// Route definition
r.Get("/{id}", handlers.GetUser)

// Inside handler — read the param
id := chi.URLParam(r, "id")
// GET /api/v1/users/42 → id = "42"
```

### 3. Route grouping
```go
r.Route("/api/v1", func(r chi.Router) {
    r.Route("/users", func(r chi.Router) {
        r.Get("/", handlers.ListUsers)
        r.Post("/", handlers.CreateUser)
    })
})
```
All routes inside inherit the prefix. Clean, readable, no repetition.

### 4. Reading request body (POST/PUT)
```go
var user models.User
json.NewDecoder(r.Body).Decode(&user)
```
`r.Body` is the raw request body stream.
`Decode(&user)` parses JSON and fills the struct. `&user` = pointer to user.

### 5. Map — in-memory store
```go
var store = map[string]models.User{
    "1": {ID: "1", Name: "Akshad", Email: "akshad@example.com"},
}

// Read
user, ok := store["1"]  // ok = false if key doesn't exist

// Write
store["3"] = newUser

// Delete
delete(store, "1")
```

### 6. HTTP status codes used
| Code | Constant | When to use |
|------|----------|-------------|
| 200 | `http.StatusOK` | Successful GET/PUT |
| 201 | `http.StatusCreated` | Successful POST (resource created) |
| 204 | `http.StatusNoContent` | Successful DELETE (no body) |
| 400 | `http.StatusBadRequest` | Bad input from client |
| 404 | `http.StatusNotFound` | Resource doesn't exist |

### 7. Middleware
```go
r.Use(middleware.Logger)    // logs every request
r.Use(middleware.Recoverer) // catches panics, returns 500
```
Middleware runs before every handler. `r.Use()` registers it globally.
We'll build our own middleware in Stage 03.

---

## How to run

```bash
# From repo root
cd stage-02-routing

# Run the server
go run main.go
```

You'll see:
```
Stage 02 — Server starting on http://localhost:8080
Routes:
  GET    /health
  GET    /api/v1/users
  POST   /api/v1/users
  GET    /api/v1/users/{id}
  PUT    /api/v1/users/{id}
  DELETE /api/v1/users/{id}
```

Chi's Logger middleware will print every request as it comes in:
```
2026/04/18 10:00:00 "GET http://localhost:8080/api/v1/users HTTP/1.1" 200 45B 120µs
```

---

## Test the endpoints

Open `requests.http` in VS Code with the REST Client extension and click `Send Request` on each one.

Or use curl:

```bash
# List users
curl http://localhost:8080/api/v1/users

# Get one user
curl http://localhost:8080/api/v1/users/1

# User not found
curl http://localhost:8080/api/v1/users/999

# Create user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Ketaki", "email": "ketaki@example.com"}'

# Update user
curl -X PUT http://localhost:8080/api/v1/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "Akshad Updated", "email": "new@example.com"}'

# Delete user
curl -X DELETE http://localhost:8080/api/v1/users/2
```

---

## What's missing (coming next)

| Missing | Added in |
|---------|----------|
| Custom middleware (request ID, CORS, auth guards) | Stage 03 — Middleware |
| Real database instead of in-memory map | Stage 04 — PostgreSQL |
| Input validation (email format, required fields) | Stage 06 — Validation |
| UUIDs instead of incremental IDs | Stage 06 — Validation |
