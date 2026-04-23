# Multi-stage Dockerfile for the Go backend
#
# Why multi-stage?
# ─────────────────
# Stage 1 (builder): Uses the full Go toolchain image (~700MB) to compile the binary.
# Stage 2 (runner):  Uses a tiny Alpine Linux image (~5MB) with ONLY the compiled binary.
#
# Result: the final Docker image is ~15MB instead of ~700MB.
# The Go toolchain, source code, and build cache never end up in production.
#
# This is the standard Go deployment pattern — Go compiles to a single static binary
# with no runtime dependencies, so the runner image needs almost nothing.

# ─────────────────────────────────────────────────────────────
# Stage 1: Build
# ─────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

# Set the working directory inside the builder container.
# All subsequent commands run relative to this path.
WORKDIR /src

# Copy go.mod and go.sum FIRST — before copying source code.
# Docker caches each layer. If go.mod/go.sum don't change, `go mod download`
# is skipped on the next build (huge time saver during development).
COPY go.mod go.sum ./

# Download all dependencies declared in go.mod.
# Runs inside the container — no internet access needed at runtime.
RUN go mod download

# Now copy the rest of the source code.
# Separated from go.mod so that dependency downloads are cached independently
# from source code changes. Only this layer re-runs when you edit .go files.
COPY . .

# Compile the binary.
# CGO_ENABLED=0  — disable C bindings (produces a fully static binary)
#                  Required for Alpine Linux (which has no glibc)
# GOOS=linux     — build for Linux even if you're on macOS or Windows
# -o /app        — output binary path
# ./stage-10-deployment/  — the package to build (main.go lives there)
RUN CGO_ENABLED=0 GOOS=linux go build -o /app ./stage-10-deployment/

# ─────────────────────────────────────────────────────────────
# Stage 2: Run
# ─────────────────────────────────────────────────────────────
FROM alpine:3.19

# Add two small but important packages:
# ca-certificates — TLS root certificates (needed for HTTPS outbound calls)
# tzdata          — timezone data (for correct time zone handling in logs)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy ONLY the compiled binary from the builder stage.
# Nothing else (no source, no Go toolchain, no build cache) comes with it.
COPY --from=builder /app .

# EXPOSE documents which port the app listens on.
# It does NOT actually publish the port — that's done in docker-compose.yml
# with the `ports:` directive. EXPOSE is just metadata / documentation.
EXPOSE 8080

# CMD is the default command when the container starts.
# Uses exec form (JSON array) — not shell form — so signals (SIGTERM) go
# directly to the process, enabling graceful shutdown.
CMD ["./app"]
