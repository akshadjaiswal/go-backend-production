# CLAUDE.md — go-backend-production

This file tells Claude how to work in this repo consistently across sessions.

---

## What this repo is

A step-by-step Go backend learning project. Each stage is self-contained and builds on the previous one conceptually. The goal is to understand how production Go backends are built — not just make things work.

**Owner:** Akshad Jaiswal (new to Go, comes from JS/Node background)

---

## Current progress

| Stage | Status | Topic |
|-------|--------|-------|
| stage-01-basics | ✅ Done | net/http, handlers, JSON |
| stage-02-routing | ✅ Done | Chi router, path params, multi-package |
| stage-03-middleware | ✅ Done | Custom middleware, context, CORS, AuthGuard |
| stage-04-database | ✅ Done | PostgreSQL, sqlx, UUIDs, migrations |
| stage-05-auth | ✅ Done | bcrypt, JWT, protected routes |
| stage-06-validation | ✅ Done | go-playground/validator, field-level errors |
| stage-07-config | ✅ Done | godotenv, Config struct, fail-fast |
| stage-08-logging | ✅ Done | Structured JSON logging with slog |
| stage-09-testing | ✅ Done | Unit + integration tests |
| stage-10-deployment | ✅ Done | Docker, multi-stage builds, docker-compose |

**Always update this table when a stage is completed.**

---

## How stages are structured

Every stage follows this exact pattern — do not deviate:

```
stage-XX-name/
├── main.go              ← entry point only, wires everything
├── config/config.go     ← from stage-07 onwards, always use config
├── db/db.go             ← from stage-04 onwards
├── migrations/*.sql     ← plain SQL, one file per migration
├── models/user.go       ← data shapes with json + db + validate tags
├── handlers/            ← one file per resource (users.go, auth.go)
├── middleware/          ← jwt.go and any custom middleware
├── routes/routes.go     ← all route wiring in one place
├── validator/           ← from stage-06 onwards
├── requests.http        ← VS Code REST Client test file, ALWAYS included
└── README.md            ← ALWAYS included, detailed explanation
```

Each stage is **independent** — has its own DB, its own `.env`, its own `go run` command.

---

## Conventions to follow always

### Go patterns used across the repo
- **Handler struct pattern** — handlers are methods on a struct, never global functions with global DB vars
  ```go
  type UsersHandler struct { DB *sqlx.DB; cfg *config.Config }
  func NewUsersHandler(db *sqlx.DB, cfg *config.Config) *UsersHandler
  ```
- **Config injection** — config is loaded once in `main()` and passed to everything via constructors
- **`writeJSON` helper** — every handler package has a private `writeJSON(w, status, data)` helper
- **Parameterized SQL** — always use `$1, $2` placeholders, never string concatenation
- **`RETURNING *`** — always use on INSERT/UPDATE to get the full row back in one query
- **`sql.ErrNoRows`** — always check this specifically before generic error → returns 404

### HTTP status codes used
| Situation | Status |
|-----------|--------|
| Success GET/PUT | 200 |
| Resource created | 201 |
| Successful DELETE | 204 |
| Bad JSON / invalid UUID | 400 |
| Auth failed | 401 |
| Validation failed | 422 |
| Duplicate email | 409 |
| Not found | 404 |
| Server error | 500 |

### Error response format
```json
{ "error": "message here" }
{ "error": "validation failed", "fields": [{"field": "email", "message": "..."}] }
```

### Database naming
Each stage gets its own PostgreSQL database:
```
go_backend_production_stage04
go_backend_production_stage05
go_backend_production_stage06
go_backend_production_stage07
...
```

---

## Tech stack (locked — don't change without reason)

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/go-chi/chi/v5` | v5.2.5 | HTTP router |
| `github.com/jmoiron/sqlx` | v1.4.0 | PostgreSQL wrapper |
| `github.com/lib/pq` | v1.12.3 | PostgreSQL driver |
| `github.com/golang-jwt/jwt/v5` | v5.3.1 | JWT tokens |
| `golang.org/x/crypto` | v0.50.0 | bcrypt |
| `github.com/go-playground/validator/v10` | v10.30.2 | Validation |
| `github.com/joho/godotenv` | v1.5.1 | .env loading |

Do not introduce new packages without discussing first. The goal is learning, not accumulating dependencies.

---

## What to include in every README

Every stage README must have:
1. **Goal** — one line explaining what this stage teaches
2. **What changed from previous stage** — comparison table
3. **Project structure** — file tree with comments
4. **Key concepts** — every new concept explained with code examples
5. **Setup** — exact commands to create DB, run migration, start server
6. **Test flow** — step by step what to test and expected responses
7. **What's missing** — table pointing to next stages

---

## What to include in every requests.http

- Health check always first
- Cover happy path AND error cases (missing fields, wrong auth, invalid IDs)
- Add comments above each request explaining what it tests
- Include `PASTE_TOKEN_HERE` placeholders where JWT is needed

---

## Commit conventions

- 3 commits per stage, split logically:
  1. Core domain (models, db, config, migrations)
  2. Business logic (handlers, middleware)
  3. Wiring + docs (routes, main, requests.http, README)
- Commit message format: `stage-XX: short description`
- Real timestamps — never backdate commits
- No `Co-Authored-By` lines

---

## Testing workflow

1. Implement full stage
2. Build check: `go build ./stage-XX-name/...`
3. Run server and test all endpoints with curl
4. Kill server + close all connections
5. Give user a testing guide in chat
6. User tests manually
7. User says commit → then commit in 3

---

## Go version

`go 1.26.2` — module path: `github.com/akshadjaiswal/go-backend-production`

---

## Things Akshad wants in every session

- Explain every concept — he's learning Go from scratch (JS/Node background)
- Heavy comments in code — every non-obvious line explained
- Don't skip explaining why, not just what
- Always plan before building — show the plan, get approval, then build
- Detailed READMEs — treat them as teaching material, not just docs
