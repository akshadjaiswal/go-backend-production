# go-backend-production

A step-by-step Go backend built for production — covering routing, auth, databases, logging, testing, and deployment.

> **Learn online →** [go-backend-production.vercel.app](https://go-backend-production.vercel.app)
> Browse all 10 stages with rendered docs, inline source code explorer, and fuzzy search.

This repo is designed for developers who are **new to Go** but want to understand how real production backends are built. Each stage is self-contained, heavily commented, and comes with its own README explaining every concept from scratch.

---

## Who is this for?

- You know JavaScript/Node/Python but want to learn Go
- You want to understand what frameworks do under the hood
- You want to build backends the right way — not just make things work

---

## Documentation site

The easiest way to learn from this repo is the interactive docs site:

**[go-backend-production.vercel.app](https://go-backend-production.vercel.app)**

- All 10 stages rendered as navigable docs pages
- **Side-by-side split view** — docs on the left, source code explorer on the right (VS Code-style: file tree, line numbers, syntax highlighting). Drag the divider to resize.
- Browse every file inline: `main.go`, handlers, middleware, migrations, `Dockerfile`, `.env.example`, and more — no need to clone
- Client-side fuzzy search across all stage content (`/` to open)
- Table of contents, reading progress, dark mode, bookmarks, keyboard shortcuts (`?` for reference)

If you prefer reading code directly, clone the repo and follow the steps below.

---

## How to use this repo

Each stage builds on the previous one conceptually, but every stage is **independent** — you can run any stage on its own without the others.

1. Open the [docs site](https://go-backend-production.vercel.app) and pick a stage
2. Read the docs on the left — the source code is live on the right in the editor pane
3. Clone and run the stage locally to experiment with the actual code
4. Test endpoints using the `requests.http` file in VS Code REST Client
5. Mark the stage complete, move to the next one

---

## Stages

| Stage | Topic | What you learn |
|-------|-------|---------------|
| [Stage 01](./stage-01-basics) | HTTP Basics | `net/http`, handlers, JSON responses, query params |
| [Stage 02](./stage-02-routing) | Routing with Chi | Path params, route groups, multi-package structure |
| [Stage 03](./stage-03-middleware) | Middleware | Custom middleware, context, CORS, API key guard |
| [Stage 04](./stage-04-database) | PostgreSQL | `sqlx`, raw SQL, connection pooling, UUIDs, migrations |
| [Stage 05](./stage-05-auth) | JWT Auth | bcrypt, JWT tokens, protected routes, Bearer auth |
| [Stage 06](./stage-06-validation) | Validation | Input validation, email format, custom error responses |
| [Stage 07](./stage-07-config) | Config | `.env` files, environment variables, config structs |
| [Stage 08](./stage-08-logging) | Logging | Structured JSON logging with `slog` |
| [Stage 09](./stage-09-testing) | Testing | Unit tests, integration tests, test DB setup |
| [Stage 10](./stage-10-deployment) | Deployment | Docker, multi-stage builds, docker-compose |

---

## Tech stack

| Tool | Purpose |
|------|---------|
| Go 1.22+ | Language |
| [Chi](https://github.com/go-chi/chi) | HTTP router |
| [sqlx](https://github.com/jmoiron/sqlx) | PostgreSQL driver wrapper |
| [golang-jwt](https://github.com/golang-jwt/jwt) | JWT tokens |
| [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) | Password hashing |
| PostgreSQL | Database |
| Docker | Containerization (Stage 10) |

---

## Prerequisites

- Go 1.22+ — [install](https://go.dev/dl/)
- PostgreSQL — `brew install postgresql@16`
- VS Code with [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) extension (for `.http` files)

---

## Quick start (Stage 01 — no DB needed)

```bash
git clone https://github.com/akshadjaiswal/go-backend-production.git
cd go-backend-production/stage-01-basics
go run main.go
```

```bash
curl http://localhost:8080/health
curl "http://localhost:8080/hello?name=Akshad"
```

---

## Project structure

```
go-backend-production/
├── go.mod                    ← single Go module for the whole repo
├── application/              ← Next.js docs site (go-backend-production.vercel.app)
├── stage-01-basics/
│   ├── main.go
│   ├── requests.http         ← test endpoints in VS Code
│   └── README.md             ← full explanation of concepts
├── stage-02-routing/
│   ├── main.go
│   ├── handlers/
│   ├── models/
│   ├── routes/
│   ├── requests.http
│   └── README.md
├── stage-03-middleware/
│   ├── middleware/           ← RequestID, Logger, CORS, AuthGuard
│   └── ...
├── stage-04-database/
│   ├── db/                   ← connection + pool setup
│   ├── migrations/           ← SQL schema files
│   └── ...
├── stage-05-auth/
│   ├── handlers/auth.go      ← register + login
│   ├── middleware/jwt.go     ← Bearer token validation
│   └── ...
└── ...
```

---

## Key Go concepts covered across all stages

- `package` and `import` system
- Structs, interfaces, and methods
- Error handling (`if err != nil`)
- HTTP handler signature `(w http.ResponseWriter, r *http.Request)`
- Middleware pattern `func(http.Handler) http.Handler`
- `context.WithValue` — passing data through request chain
- Connection pooling with `database/sql`
- Parameterized SQL queries (preventing SQL injection)
- JWT signing and verification
- bcrypt password hashing
- Dependency injection via handler structs

---

## Running tests for `.http` files

Install the **REST Client** extension in VS Code. Open any `requests.http` file — you'll see a `Send Request` button above each request. Click it to send and see the response in a split panel.

---

## Author

**Akshad Jaiswal** — building this to learn Go properly, one stage at a time.
