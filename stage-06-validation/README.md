# Stage 06 — Validation

> **Goal:** Add proper input validation to every endpoint using `go-playground/validator`. Stop bad data before it hits the database. Return clear, field-level error messages instead of cryptic DB errors.

---

## What changed from Stage 05?

| Stage 05 | Stage 06 |
|----------|----------|
| Manual `if name == ""` checks | Declarative `validate` struct tags |
| Generic "required" errors | Field-level messages: `"email: must be a valid email address"` |
| No min/max length | `min=2`, `max=100`, `min=8` enforced |
| Invalid UUID hits DB and crashes | UUID validated before DB query |
| Multiple fields fail silently | All failing fields returned at once |

---

## Project structure

```
stage-06-validation/
├── main.go
├── db/db.go
├── migrations/001_create_users.sql
├── models/
│   └── user.go              ← validate tags on all request structs
├── handlers/
│   ├── auth.go              ← validate before processing
│   └── users.go             ← validate body + UUID path params
├── middleware/jwt.go
├── routes/routes.go
├── validator/
│   └── validator.go         ← shared validator + error formatter
├── requests.http
└── README.md
```

---

## Key concepts

### 1. Validation tags on structs

```go
type RegisterRequest struct {
    Name     string `json:"name"     validate:"required,min=2,max=100"`
    Email    string `json:"email"    validate:"required,email"`
    Password string `json:"password" validate:"required,min=8,max=72"`
}
```

The `validate` tag works exactly like `json` and `db` tags — the validator library reads them at runtime.

Common tags:

| Tag | Meaning |
|-----|---------|
| `required` | Field must be present and non-empty |
| `email` | Must be valid email format (has `@` and domain) |
| `min=N` | String length ≥ N |
| `max=N` | String length ≤ N |
| `uuid4` | Must be a valid UUID v4 |
| `oneof=a b c` | Must be one of the listed values |

### 2. Validating a struct

```go
if err := v.Validate.Struct(req); err != nil {
    writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
        "error":  "validation failed",
        "fields": v.FormatErrors(err),
    })
    return
}
```

`422 Unprocessable Entity` is the correct HTTP status for validation errors — the request was well-formed JSON but the data was invalid.

### 3. Validating a single value (path params)

```go
id := chi.URLParam(r, "id")

// Validate just a single variable, not a struct
if err := v.Validate.Var(id, "required,uuid4"); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user ID format"})
    return
}
```

`Validate.Var()` validates a single value against a tag string — useful for path params.

### 4. Readable error responses

Without custom formatting, validator returns:
```
Key: 'RegisterRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag
```

With our `FormatErrors` helper:
```json
{
  "error": "validation failed",
  "fields": [
    {"field": "email", "message": "must be a valid email address"},
    {"field": "password", "message": "minimum length is 8 characters"}
  ]
}
```

All failing fields are returned at once — the client doesn't have to fix and resubmit one error at a time.

### 5. Single validator instance

```go
// validator/validator.go
var Validate = validator.New()
```

We create one validator and reuse it everywhere. It caches struct reflection data internally — creating a new one per request would be wasteful.

### 6. Why max=72 for passwords?

bcrypt silently truncates passwords longer than 72 bytes. If someone sets a 100-char password, bcrypt only hashes the first 72 chars. By enforcing `max=72`, we avoid this silent truncation.

---

## HTTP status codes for errors

| Situation | Status | Code |
|-----------|--------|------|
| Malformed JSON | 400 Bad Request | Invalid body |
| Validation failed | 422 Unprocessable Entity | Field rules failed |
| Wrong credentials | 401 Unauthorized | Auth failed |
| Resource not found | 404 Not Found | No DB row |
| Duplicate email | 409 Conflict | Unique constraint |
| DB/server error | 500 Internal Server Error | Unexpected |

---

## Setup

```bash
# Create DB
createdb go_backend_production_stage06

# Run migration
psql -d go_backend_production_stage06 -f migrations/001_create_users.sql

# Start server
cd stage-06-validation
DATABASE_URL="postgres://$(whoami)@localhost:5432/go_backend_production_stage06?sslmode=disable" go run main.go
```

---

## Test flow

1. Try registering with bad data — see field-level errors
2. Register with valid data
3. Login → copy token
4. Try hitting `/api/v1/users/not-a-uuid` — see UUID validation error
5. Try creating a user with empty name and bad email — see multiple errors at once
6. All valid requests still work normally

Open `requests.http` in VS Code and go through each request from top to bottom.

---

## Example error responses

**Missing required field:**
```json
{
  "error": "validation failed",
  "fields": [{"field": "name", "message": "Name is required"}]
}
```

**Multiple errors at once:**
```json
{
  "error": "validation failed",
  "fields": [
    {"field": "name", "message": "minimum length is 2 characters"},
    {"field": "email", "message": "must be a valid email address"},
    {"field": "password", "message": "minimum length is 8 characters"}
  ]
}
```

**Invalid UUID path param:**
```json
{"error": "invalid user ID format"}
```

---

## What's missing (coming next)

| Missing | Added in |
|---------|----------|
| DB URL hardcoded / from env var | Stage 07 — Config |
| JWT secret hardcoded | Stage 07 — Config |
| No structured logging | Stage 08 — Logging |
