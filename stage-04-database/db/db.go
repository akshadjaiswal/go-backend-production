// package db handles the database connection setup.
// The rest of the app imports this package to get a *sqlx.DB instance.
package db

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	// _ "github.com/lib/pq" — the underscore import is a blank import.
	// We don't use pq directly, but importing it registers the "postgres"
	// driver with database/sql under the hood. Without this, sql.Open("postgres", ...)
	// would fail with "unknown driver".
	_ "github.com/lib/pq"
)

// Connect opens a connection to PostgreSQL and verifies it with a ping.
// Returns a *sqlx.DB — the connection pool we'll use across the entire app.
//
// dsn = Data Source Name — the connection string format for PostgreSQL:
//   "postgres://user:password@host:port/dbname?sslmode=disable"
func Connect(dsn string) (*sqlx.DB, error) {
	// sqlx.Open doesn't actually connect — it just validates the DSN format.
	// The real connection happens on the first query or on db.Ping().
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	// Connection pool settings — important for production:
	// MaxOpenConns: max simultaneous connections to PostgreSQL
	// MaxIdleConns: connections kept open when idle (reused for next request)
	// ConnMaxLifetime: how long a connection can live before being replaced
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Ping actually opens a connection and verifies the DB is reachable.
	// Fail fast at startup rather than failing on the first request.
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	fmt.Println("Connected to PostgreSQL")
	return db, nil
}
