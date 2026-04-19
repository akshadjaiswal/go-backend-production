# Stage 04 — PostgreSQL Database

> **Goal:** Replace the in-memory map with a real PostgreSQL database using `sqlx`. Learn raw SQL in Go, connection pooling, struct scanning, and proper error handling.

---

## What changed from Stage 03?

| Stage 03 | Stage 04 |
|----------|----------|
| `map[string]User` in memory | PostgreSQL database |
| Data lost on restart | Data persists |
| String IDs (`"1"`, `"2"`) | UUID primary keys |
| No timestamps | `created_at`, `updated_at` columns |
| Handler functions | Handler struct with DB injected |

---

## Why `sqlx` over GORM?

`sqlx` is a thin wrapper over Go's standard `database/sql`. You write real SQL — no magic, no hidden queries.

```go
// sqlx — you see exactly what runs
db.Select(&users, "SELECT id, name, email FROM users ORDER BY created_at DESC")

// GORM — hides the SQL
db.Find(&users)
```

Learning `sqlx` first means you understand what's actually happening in the database. You can always add an ORM later, but you can't un-learn bad habits.

---

## Project structure

```
stage-04-database/
├── main.go                       ← connects to DB, wires everything together
├── db/
│   └── db.go                     ← connection setup + pool config
├── migrations/
│   └── 001_create_users.sql      ← schema — run this first
├── models/
│   └── user.go                   ← User struct with json + db tags
├── handlers/
│   └── users.go                  ← handler struct pattern, real SQL queries
├── routes/
│   └── routes.go
├── middleware/
│   └── auth.go
├── requests.http
└── README.md
```

---

## Prerequisites — set up PostgreSQL

### 1. Install PostgreSQL (if not already)
```bash
brew install postgresql@16
brew services start postgresql@16
```

### 2. Create the database
```bash
createdb go_backend_production
```

### 3. Run the migration
```bash
psql -d go_backend_production -f migrations/001_create_users.sql
```

You should see:
```
CREATE EXTENSION
CREATE TABLE
INSERT 0 2
```

---

## How to run

```bash
cd stage-04-database

# Option 1: use the default DSN (postgres://postgres:postgres@localhost:5432/go_backend_production)
go run main.go

# Option 2: set a custom DSN via environment variable
DATABASE_URL="postgres://youruser:yourpass@localhost:5432/go_backend_production?sslmode=disable" go run main.go
```

---

## Key concepts

### 1. Handler struct pattern

Instead of global variables, we inject dependencies into a struct:

```go
type UsersHandler struct {
    DB *sqlx.DB
}

func NewUsersHandler(db *sqlx.DB) *UsersHandler {
    return &UsersHandler{DB: db}
}

// Methods on the struct are our handlers
func (h *UsersHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
    // h.DB is available here
}
```

**Why?** Testable — you can pass a test DB. No global state.

### 2. `db` struct tags

```go
type User struct {
    ID    string `json:"id"    db:"id"`
    Name  string `json:"name"  db:"name"`
    Email string `json:"email" db:"email"`
}
```

`db:"id"` tells sqlx: "when scanning a row, put the `id` column into this field."

### 3. `db.Select` vs `db.Get` vs `db.Exec`

| Method | Use for | Returns |
|--------|---------|---------|
| `db.Select(&slice, query)` | Multiple rows | Fills a slice |
| `db.Get(&struct, query)` | Single row | Fills one struct, returns `sql.ErrNoRows` if not found |
| `db.Exec(query)` | INSERT/UPDATE/DELETE without needing result | `sql.Result` (rows affected) |

### 4. Parameterized queries — never concatenate SQL

```go
// WRONG — SQL injection vulnerability
query := "SELECT * FROM users WHERE id = '" + id + "'"

// CORRECT — $1 is a placeholder, pq fills it safely
db.Get(&user, "SELECT * FROM users WHERE id = $1", id)
```

PostgreSQL uses `$1, $2, $3...` for placeholders. MySQL uses `?`.

### 5. `sql.ErrNoRows`

```go
err := db.Get(&user, "SELECT ... WHERE id = $1", id)
if err == sql.ErrNoRows {
    // 404 — row doesn't exist
}
if err != nil {
    // 500 — something else went wrong
}
```

Always check `sql.ErrNoRows` before a generic error check.

### 6. `RETURNING` in PostgreSQL

```sql
INSERT INTO users (name, email)
VALUES ($1, $2)
RETURNING id, name, email, created_at, updated_at
```

`RETURNING` makes PostgreSQL return the inserted/updated row immediately. Perfect with `db.Get` — one query instead of insert + select.

### 7. Connection pool

```go
db.SetMaxOpenConns(25)       // max 25 simultaneous DB connections
db.SetMaxIdleConns(5)        // keep 5 connections warm when idle
db.SetConnMaxLifetime(5 * time.Minute) // recycle connections every 5 min
```

Go's `database/sql` is a connection pool by default. You don't open/close connections per request — the pool manages it.

### 8. Blank import

```go
import _ "github.com/lib/pq"
```

`_` means "import for side effects only". `pq` registers itself as the `"postgres"` driver when imported. Without this, `sqlx.Open("postgres", ...)` would fail.

---

## UUID vs auto-increment

```sql
id UUID PRIMARY KEY DEFAULT uuid_generate_v4()
```

vs

```sql
id SERIAL PRIMARY KEY
```

| | UUID | SERIAL |
|-|------|--------|
| Guessable | No | Yes (`/users/1`, `/users/2`...) |
| Distributed safe | Yes | No |
| URL looks like | `/users/550e8400-e29b-41d4-a716-446655440000` | `/users/1` |
| Production use | Preferred | Fine for small apps |

---

## Test the endpoints

Open `requests.http` in VS Code. First run `List all users` to get real UUIDs, then replace `REPLACE_WITH_REAL_UUID` in the other requests.

```bash
# List users
curl http://localhost:8080/api/v1/users -H "X-API-Key: secret-key-123"

# Create user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -H "X-API-Key: secret-key-123" \
  -d '{"name": "Ketaki", "email": "ketaki@example.com"}'

# Duplicate email → 409
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -H "X-API-Key: secret-key-123" \
  -d '{"name": "Dup", "email": "akshad@example.com"}'
```

---

## What's missing (coming next)

| Missing | Added in |
|---------|----------|
| JWT auth (login, tokens) | Stage 05 — Auth |
| Input validation (email format, length) | Stage 06 — Validation |
| DB URL from .env file | Stage 07 — Config |
| Structured JSON logging | Stage 08 — Logging |
