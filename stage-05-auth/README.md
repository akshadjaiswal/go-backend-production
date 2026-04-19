# Stage 05 — JWT Authentication

> **Goal:** Replace the hardcoded API key with proper JWT authentication. Users register, login, and receive a token. Protected routes require `Authorization: Bearer <token>`.

---

## What changed from Stage 04?

| Stage 04 | Stage 05 |
|----------|----------|
| `X-API-Key: secret-key-123` | `Authorization: Bearer <jwt>` |
| Hardcoded key | Token per user, expires in 24h |
| No user accounts | Register + Login endpoints |
| No passwords | bcrypt-hashed passwords in DB |

---

## How JWT auth works (the full flow)

```
1. Register
   POST /auth/register {"name", "email", "password"}
   → bcrypt hashes the password
   → stores user + hash in DB
   → returns user (no token)

2. Login
   POST /auth/login {"email", "password"}
   → fetches user from DB
   → bcrypt compares password with hash
   → generates JWT: {user_id, email, exp: now+24h}
   → returns {"token": "eyJ...", "user": {...}}

3. Use protected routes
   GET /api/v1/users
   Authorization: Bearer eyJ...
   → JWTAuth middleware validates token
   → attaches user_id to context
   → handler runs
```

---

## Project structure

```
stage-05-auth/
├── main.go
├── db/db.go
├── migrations/
│   ├── 001_create_users.sql     ← includes password_hash column
│   └── 002_add_password_hash.sql ← if upgrading from stage-04 DB
├── models/user.go               ← User with json:"-" on password_hash
├── handlers/
│   ├── auth.go                  ← Register + Login
│   └── users.go                 ← CRUD (same as before)
├── middleware/
│   └── jwt.go                   ← validates Bearer token
├── routes/routes.go
├── requests.http
└── README.md
```

---

## Key concepts

### 1. bcrypt — password hashing

```go
// Hash (on register)
hash, _ := bcrypt.GenerateFromPassword([]byte("mypassword"), bcrypt.DefaultCost)
// hash = "$2a$10$..." — includes the salt, looks different every time

// Compare (on login)
err := bcrypt.CompareHashAndPassword(hash, []byte("mypassword"))
// err == nil → passwords match
// err != nil → wrong password
```

**Never store plain text passwords.** bcrypt is:
- One-way (can't reverse the hash)
- Salted automatically (same password → different hash each time)
- Slow by design (makes brute force hard)

### 2. JWT structure

A JWT looks like: `eyJhbGciOiJIUzI1NiJ9.eyJ1c2VyX2lkIjoiMTIzIn0.abc123`

It has 3 parts separated by `.`:
```
base64(header) . base64(payload) . signature
```

- **Header** — algorithm used (`HS256`)
- **Payload** — your data (`user_id`, `email`, `exp`)
- **Signature** — HMAC of header+payload using your secret key

The payload is NOT encrypted — anyone can decode it. The signature ensures it wasn't tampered with.

### 3. JWT claims

```go
claims := jwt.MapClaims{
    "user_id": user.ID,
    "email":   user.Email,
    "exp":     time.Now().Add(24 * time.Hour).Unix(), // expiry
    "iat":     time.Now().Unix(),                      // issued at
}
```

Standard claims: `exp` (expiry), `iat` (issued at), `sub` (subject), `iss` (issuer).

### 4. JWT middleware — validating the token

```go
token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    // Verify algorithm — prevents algorithm confusion attacks
    if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, jwt.ErrSignatureInvalid
    }
    return jwtSecret, nil
})
```

`jwt.Parse` automatically:
- Decodes the token
- Verifies the signature
- Checks if it's expired

### 5. `json:"-"` — never expose sensitive fields

```go
type User struct {
    ID           string `json:"id"`
    PasswordHash string `json:"-"` // ← never included in JSON output
}
```

The `-` tag means: skip this field when marshaling to JSON. Even if you accidentally pass a User with a password hash to `json.Encode`, it won't appear in the response.

### 6. Same error for wrong email and wrong password

```go
// DON'T do this — reveals whether email exists
if user not found → "user not found"
if wrong password → "wrong password"

// DO this — attacker can't enumerate emails
if user not found OR wrong password → "invalid email or password"
```

---

## Setup

### 1. Create a fresh database for this stage
```bash
createdb go_backend_production_stage05
```

### 2. Run the migration
```bash
psql -d go_backend_production_stage05 -f migrations/001_create_users.sql
```

### 3. Start the server
```bash
cd stage-05-auth
DATABASE_URL="postgres://$(whoami)@localhost:5432/go_backend_production_stage05?sslmode=disable" go run main.go
```

---

## Test flow

1. **Register** — `POST /auth/register`
2. **Login** — `POST /auth/login` → copy the `token` from the response
3. **Use token** — paste into `Authorization: Bearer <token>` in VS Code REST Client
4. Try without token → 401
5. Try with wrong token → 401

Open `requests.http` in VS Code and follow from top to bottom.

---

## What's missing (coming next)

| Missing | Added in |
|---------|----------|
| Input validation (email format, password strength) | Stage 06 — Validation |
| JWT secret from env var | Stage 07 — Config |
| Refresh tokens | Beyond scope (production pattern) |
