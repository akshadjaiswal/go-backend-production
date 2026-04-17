# Stage 01 — Go Basics: HTTP Server

> **Goal:** Spin up a working HTTP server using only Go's standard library. No external packages. Understand how Go is structured from the ground up.

---

## What you'll learn

- How a Go program is structured (`package`, `import`, `func main`)
- What a `struct` is and how JSON tags work
- How Go's built-in HTTP server works (`net/http`)
- What a handler function is
- How to read query parameters from a URL
- How to write JSON responses

---

## Why standard library first?

In JS/Node land you'd immediately reach for Express. In Go, the standard library's `net/http` is powerful enough that many production companies use it directly. Starting here means you understand **what frameworks are doing under the hood** before you use them.

---

## Go concepts introduced

| Concept | What it is | JS/Node equivalent |
|---------|-----------|-------------------|
| `package main` | Entry point package | `index.js` / `main.js` |
| `import` | Bring in other packages | `require` / `import` |
| `func` | Define a function | `function` / `const fn =` |
| `struct` | Custom data type with fields | Class / interface in TS |
| `json:"field"` | JSON field name tag | `@JsonProperty` / nothing needed in JS |
| `http.ServeMux` | Router — maps paths to handlers | Express `app` / router |
| `http.ResponseWriter` | Write response headers + body | `res` in Express |
| `*http.Request` | Everything about the incoming request | `req` in Express |
| `:=` | Declare and assign a variable | `const` / `let` |
| `fmt.Sprintf` | String formatting | Template literals (backticks) |

---

## File structure

```
stage-01-basics/
├── main.go     ← the entire server (single file for now)
└── README.md   ← this file
```

---

## How to run

> Make sure you're in the `stage-01-basics` directory.

```bash
# From the repo root:
cd stage-01-basics

# Run the server
go run main.go
```

You should see:
```
Server starting on http://localhost:8080
Press Ctrl+C to stop
```

The server is now running. Open a **new terminal tab** to test it (or use a REST client like Postman/Insomnia).

---

## Test the endpoints

### Health check
```bash
curl http://localhost:8080/health
```
Expected response:
```json
{"message":"Server is running","status":200}
```

### Hello with name
```bash
curl "http://localhost:8080/hello?name=Akshad"
```
Expected response:
```json
{"message":"Hello, Akshad!","status":200}
```

### Hello without name (uses default)
```bash
curl http://localhost:8080/hello
```
Expected response:
```json
{"message":"Hello, World!","status":200}
```

---

## How the code flows

```
go run main.go
    → main() runs
    → Creates a mux (router)
    → Registers routes: GET /health → healthHandler, GET /hello → helloHandler
    → Starts server on :8080 (blocks here)

Incoming request: GET /health
    → mux matches route
    → calls healthHandler(w, r)
    → sets Content-Type header
    → sets 200 status code
    → encodes Response struct as JSON → writes to response
    → client receives: {"message":"Server is running","status":200}
```

---

## Key Go quirks to know

**1. Error handling is explicit**
Go doesn't have try/catch. Functions return errors as values:
```go
if err := http.ListenAndServe(":8080", mux); err != nil {
    panic(err)  // something went wrong, crash hard
}
```
This will become very natural. Every function that can fail returns `(result, error)`.

**2. `:=` vs `=`**
```go
name := "Akshad"   // declare new variable + assign (short form)
name = "Sid"       // reassign existing variable
```

**3. Unused imports = compile error**
Go won't compile if you import something you don't use. This keeps code clean.

**4. Capital letter = exported (public)**
```go
type Response struct { ... }   // capital R = usable outside this package
type response struct { ... }   // lowercase r = private to this package
```

---

## What's missing (coming in next stages)

| Missing | Added in |
|---------|----------|
| Better routing (path params like `/users/:id`) | Stage 02 — Routing with Chi |
| Request logging | Stage 03 — Middleware |
| Database | Stage 04 — PostgreSQL |
| Authentication | Stage 05 — JWT Auth |
| Input validation | Stage 06 — Validation |
| Config/env vars | Stage 07 — Config |
| Structured logging | Stage 08 — Logging |
| Tests | Stage 09 — Testing |
| Docker + deployment | Stage 10 — Deployment |

---

## Stop the server

Press `Ctrl+C` in the terminal where the server is running.
