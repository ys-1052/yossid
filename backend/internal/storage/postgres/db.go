package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ys-1052/yossid/backend/internal/config"
	"github.com/ys-1052/yossid/backend/internal/storage/postgres/db"
)

type DB struct {
	*sql.DB
	Queries *db.Queries
}

// NewDB initializes a PostgreSQL connection pool and returns a DB wrapper.
func NewDB(cfg *config.Config) (*DB, error) {
	conn, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Connection pooling configuration as requested in the design specs
	conn.SetMaxOpenConns(5)
	conn.SetMaxIdleConns(1)
	conn.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	queries := db.New(conn)

	return &DB{
		DB:      conn,
		Queries: queries,
	}, nil
}

// ExecuteTx runs a function in a database transaction.
func (d *DB) ExecuteTx(ctx context.Context, fn func(*db.Queries) error) error {
	tx, err := d.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := db.New(tx)
	if err := fn(q); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %v, rollback error: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}
