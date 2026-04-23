// package db handles PostgreSQL connection setup and pooling.
//
// In Docker, the DATABASE_URL uses the service name as the host:
//   postgres://goapp:secret@postgres:5432/go_backend_production_stage10?sslmode=disable
//                                  ^^^^^^^^
//                        Docker's internal DNS resolves "postgres" to the
//                        postgres container's IP automatically.
//
// This is docker-compose networking: containers in the same compose file
// can reach each other by service name. No IPs needed.
package db

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver — side-effect import registers the "postgres" driver
)

// Connect opens a PostgreSQL connection pool and verifies it with a ping.
// Returns an error if the DB is unreachable — caller should exit(1).
//
// Connection pool settings:
//   - MaxOpenConns: max simultaneous connections to PostgreSQL
//   - MaxIdleConns: connections kept open even when not in use (avoids reconnect overhead)
//   - ConnMaxLifetime: how long before a connection is recycled (avoids stale connections)
func Connect(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	// These settings mirror what you'd use in production.
	// In Docker, the postgres container must be healthy before the app starts
	// (enforced by depends_on + healthcheck in docker-compose.yml).
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Ping actually opens a connection and tests it.
	// sqlx.Open() is lazy — it doesn't connect until you use the DB.
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	slog.Debug("database connection established",
		slog.Int("max_open_conns", 25),
		slog.Int("max_idle_conns", 5),
	)
	return db, nil
}
