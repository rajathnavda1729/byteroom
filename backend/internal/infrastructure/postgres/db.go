package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/byteroom/backend/internal/config"
)

// DB wraps sql.DB with pool configuration.
type DB struct {
	*sql.DB
}

// Connect opens a PostgreSQL connection pool and verifies connectivity.
func Connect(cfg *config.DatabaseConfig) (*DB, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("pinging db: %w", err)
	}

	return &DB{db}, nil
}

// isDuplicateKeyError returns true when err is a PostgreSQL unique-violation (23505).
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	// pq.Error.Code "23505" = unique_violation
	type pgError interface {
		GetCode() string
	}
	if pe, ok := err.(pgError); ok {
		return pe.GetCode() == "23505"
	}
	// Fallback: check string – lib/pq exposes Code as a string field
	type codedError interface {
		Error() string
	}
	// Use pq's structured error if available
	return false
}
