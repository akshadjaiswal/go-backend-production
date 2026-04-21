# Stage 09 — Testing

> **Goal:** Add automated tests to the backend — unit tests for isolated logic, integration tests against a real database. Every handler, every validation path, and every error case is covered. No more "it worked when I manually tested it."

---

## What changed from Stage 08?

| Stage 08 | Stage 09 |
|----------|----------|
| No tests — manual testing only | 31 automated tests |
| Must run the server to verify anything | `go test ./...` runs everything in seconds |
| No test database | Separate test DB (`..._test`) |
| External test tool required (`.http` files + VS Code) | Built-in `testing` package + `go test` CLI |
| Bugs discovered during manual QA | Bugs caught the moment code is written |

---

## What structured tests look like

```
$ go test -v ./...

=== RUN   TestRegister/valid_registration
--- PASS: TestRegister/valid_registration (0.07s)
=== RUN   TestRegister/missing_name
--- PASS: TestRegister/missing_name (0.01s)
=== RUN   TestGetUser/non-existent_user_—_returns_404
--- PASS: TestGetUser/non-existent_user_—_returns_404 (0.00s)
...
ok  handlers    1.161s  coverage: 77.8% of statements
ok  middleware  0.952s  coverage: 53.1% of statements
ok  validator   1.374s  coverage: 85.7% of statements
```

---

## Project structure

```
stage-09-testing/
├── main.go                       ← unchanged from stage-08
├── config/config.go              ← unchanged
├── db/db.go                      ← unchanged
├── logger/logger.go              ← unchanged
├── models/user.go                ← unchanged
├── validator/
│   ├── validator.go              ← unchanged
│   └── validator_test.go         ← NEW: unit tests for FormatErrors
├── middleware/
│   ├── jwt.go                    ← unchanged
│   ├── logger.go                 ← unchanged
│   └── jwt_test.go               ← NEW: unit tests for JWT Handler
├── handlers/
│   ├── auth.go                   ← unchanged
│   ├── users.go                  ← unchanged
│   ├── auth_test.go              ← NEW: Register + Login integration tests
│   └── users_test.go             ← NEW: ListUsers/CreateUser/GetUser/UpdateUser/DeleteUser tests
├── testhelpers/
│   └── testhelpers.go            ← NEW: shared DB setup, cleanup, token + request helpers
├── migrations/
│   └── 001_create_users.sql      ← unchanged
├── routes/routes.go              ← unchanged
├── .env.example                  ← server config template
├── .env.test.example             ← NEW: test DB config template
└── requests.http                 ← manual testing (still useful alongside automated tests)
```

**5 new Go files.** Everything else is the same as stage-08.

---

## Key concepts

### 1. `_test.go` files — never in your binary

Any file ending in `_test.go` is only compiled when you run `go test`. It is **completely excluded** from the final binary. So tests can import debugging helpers, fake data generators, and test frameworks without bloating your production executable.

```
go build ./...     ← does NOT include _test.go files
go test ./...      ← DOES include _test.go files
```

### 2. `func TestXxx(t *testing.T)` — the one rule

Go's test runner automatically discovers functions that:
- Start with `Test`
- Take exactly one argument: `*testing.T`

```go
func TestRegister(t *testing.T) { ... }     // ✅ discovered automatically
func testHelper(t *testing.T) { ... }       // ✅ NOT discovered (lowercase)
func TestNoArgs() { ... }                    // ✅ NOT discovered (wrong signature)
```

No test framework, no annotations — just this naming convention.

### 3. `httptest` — testing HTTP handlers without a running server

This is the key package for testing HTTP in Go. It gives you two things:

```go
// 1. A fake http.ResponseWriter that captures what your handler writes
rec := httptest.NewRecorder()
// After handler runs:
rec.Code          // → the status code (default 200)
rec.Body.String() // → the response body

// 2. A real *http.Request struct (no network involved)
req := httptest.NewRequest("POST", "/auth/register", body)
req.Header.Set("Content-Type", "application/json")
```

Your handler can't tell the difference between a real request and a test request — it just sees `http.ResponseWriter` and `*http.Request`. This is what makes it testable.

```go
// Testing a handler — no server needed
rec := httptest.NewRecorder()
req := httptest.NewRequest("GET", "/health", nil)

myHandler(rec, req)   // runs synchronously

assert.Equal(t, 200, rec.Code)
```

### 4. Table-driven tests — the Go idiom

Instead of writing one `TestRegister_Valid`, one `TestRegister_MissingName`, one `TestRegister_BadEmail`... you write one function with a table:

```go
func TestRegister(t *testing.T) {
    tests := []struct {
        name           string
        body           map[string]any
        expectedStatus int
    }{
        {"valid", map[string]any{"name": "Akshad", "email": "a@b.com", "password": "pass123"}, 201},
        {"missing name", map[string]any{"email": "a@b.com", "password": "pass123"}, 422},
        {"bad email", map[string]any{"name": "Akshad", "email": "bad", "password": "pass123"}, 422},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            // run the test for this case
        })
    }
}
```

Why is this better?
- **All cases in one place** — easy to scan what's covered
- **Adding a case = one line** — no new function, no copy-paste
- **All cases run even if one fails** — you see all failures at once
- **Output shows case names**: `TestRegister/missing_name` — easy to find in CI logs

### 5. `t.Run()` subtests

`t.Run("name", func(t *testing.T) {...})` creates a **subtest**. Each subtest:
- Gets its own `*testing.T` — can fail independently
- Has a path in the output: `TestRegister/valid_registration`
- Can be run in isolation: `go test -run TestRegister/valid`

```
--- PASS: TestRegister (0.08s)
    --- PASS: TestRegister/valid_registration (0.07s)
    --- PASS: TestRegister/missing_name (0.01s)
    --- FAIL: TestRegister/invalid_email (0.00s)
        auth_test.go:82: expected 422, got 400
```

### 6. `assert` vs `require` — when to stop vs continue

Both come from `github.com/stretchr/testify`:

```go
// assert.Equal — marks test as failed but CONTINUES running
assert.Equal(t, 201, rec.Code)        // fails? → logs error, continues
assert.NotEmpty(t, body["id"])        // this still runs

// require.Equal — marks test as failed and STOPS immediately
require.Equal(t, 201, rec.Code)       // fails? → stops here, nothing below runs
assert.NotEmpty(t, body["id"])        // this does NOT run if require above failed
```

**When to use which:**
- `require` when the rest of the test can't make sense if this fails
  - `require.NoError(t, err)` after JSON decode — if decode failed, body is garbage
  - `require.Equal(t, 201, rec.Code)` in setup steps — if setup failed, test is pointless
- `assert` for the actual assertions you're checking — let all of them run so you see all failures

### 7. Test database — never corrupt your dev data

Integration tests write to a real database. We use a **completely separate database** for tests:

```
go_backend_production_stage09      ← your dev DB (server + manual testing)
go_backend_production_stage09_test ← test DB (automated tests only)
```

Tests clean up before each run with `TRUNCATE users CASCADE`. If a test crashes, the DB stays dirty — but the NEXT test cleans it up at the start, so reruns always work.

### 8. `TestMain` — package-level setup

`TestMain(m *testing.M)` runs once before ALL tests in a package:

```go
func TestMain(m *testing.M) {
    // runs BEFORE any TestXxx function
    db = connectToTestDB()

    exitCode := m.Run()   // ← this runs all the TestXxx functions

    // runs AFTER all tests
    db.Close()

    os.Exit(exitCode)  // required — signals pass/fail to the test runner
}
```

We use this to:
- Connect to the test DB once (instead of reconnecting in every test)
- Run migrations once (instead of per-test)
- Close connections cleanly after all tests finish

### 9. `testhelpers` package — shared test utilities

Test code can have its own helpers. The `testhelpers` package provides:

```go
// SetupTestDB — connects + migrates
db := testhelpers.SetupTestDB(t)

// CleanupDB — TRUNCATE users (call at start of each test)
testhelpers.CleanupDB(t, db)

// MakeTestConfig — config pointing at test DB with known JWT secret
cfg := testhelpers.MakeTestConfig()

// MakeAuthToken — returns a valid JWT string (no HTTP needed)
token := testhelpers.MakeAuthToken(t, cfg)

// NewRequest — builds httptest.Request with JSON body + optional auth header
req := testhelpers.NewRequest(t, "POST", "/auth/register", body, token)
```

Why is this not a `_test.go` file? Because `_test.go` files are package-scoped — `handlers/auth_test.go` can't import from `handlers/some_helper_test.go`. A separate non-test package (`testhelpers`) is importable by any test file in any package.

---

## Setup

### 1. Create databases

```bash
# Dev DB (for running the server)
createdb go_backend_production_stage09
psql -d go_backend_production_stage09 -f migrations/001_create_users.sql

# Test DB (for go test)
createdb go_backend_production_stage09_test
```

### 2. Create config files

```bash
# Server config
cp .env.example .env
# Edit .env — replace "youruser" with your Mac username (run: whoami)

# Test config
cp .env.test.example .env.test
# Edit .env.test — replace "youruser" with your Mac username
```

### 3. Run the tests

```bash
cd stage-09-testing

# All tests
go test ./...

# Verbose — see every subtest name
go test -v ./...

# With coverage
go test -cover ./...

# Single package
go test -v ./handlers/...

# Single test function
go test -v -run TestRegister ./handlers/...
go test -v -run TestRegister/valid_registration ./handlers/...
```

### 4. Start the server (optional — for manual testing)

```bash
go run main.go
```

---

## Test coverage explained

```
handlers    77.8% — all happy paths + common errors. Uncovered: 500 DB errors (hard to trigger)
middleware  53.1% — JWT paths covered. Uncovered: logger.go (no tests for logging middleware)
validator   85.7% — all tags covered. Uncovered: fieldMessage default case (never hit in practice)
```

**Is 100% coverage the goal?** No. 100% means every line executed — but it doesn't mean every *case* tested. A better goal is "test all business logic paths" which we've done here.

---

## Testing guide — what to run

| Command | What it does |
|---------|-------------|
| `go test ./...` | Run all tests, minimal output |
| `go test -v ./...` | Verbose — see each subtest pass/fail |
| `go test -cover ./...` | Show coverage % per package |
| `go test -v -run TestRegister` | Run only Register tests |
| `go test -v -run TestRegister/valid` | Run only the "valid" subtest |
| `go test -count=1 ./...` | Force re-run (bypass test cache) |

---

## What's missing (coming next)

| Missing | Added in |
|---------|----------|
| Docker / containerization | Stage 10 — Deployment |
| CI/CD pipeline (GitHub Actions) | Stage 10 — Deployment |
