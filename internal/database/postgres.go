package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"wednesday.wtf/godin/internal/config"
)

type Database struct {
	log    *zap.Logger
	db     *sql.DB
	config *config.Config
}

func (d *Database) clearExpiredSessions(ctx context.Context) error {
	_, err := d.db.ExecContext(ctx, `
		UPDATE eris.sessions
		SET status = 'expired', updated_at = NOW()
		WHERE status = 'open'
		AND expiration_time < NOW()
	`)
	if err == nil {
		return nil
	}

	// Older development databases may not have the status/updated_at columns yet.
	// Fall back to deleting known expired session rows from the one owned table rather
	// than dynamically scanning arbitrary schemas.
	if _, fallbackErr := d.db.ExecContext(ctx, `DELETE FROM eris.sessions WHERE expiration_time < NOW()`); fallbackErr != nil {
		return errors.Join(err, fallbackErr)
	}

	return nil
}

func New(log *zap.Logger, cfg *config.Config) (*Database, error) {
	db, err := sql.Open("postgres", cfg.PostgresConnectionString)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	database := Database{
		log:    log,
		db:     db,
		config: cfg,
	}

	if err := database.clearExpiredSessions(context.Background()); err != nil {
		log.Warn("failed to clear expired sessions", zap.Error(err))
	}

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			if err := database.clearExpiredSessions(context.Background()); err != nil {
				log.Warn("failed to clear expired sessions", zap.Error(err))
			}
		}
	}()

	return &database, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

type TransactionFn func(d *Database, tx *sql.Tx) error

func (d *Database) Transaction(fn TransactionFn) error {
	return d.TransactionContext(context.Background(), fn)
}

func (d *Database) TransactionContext(ctx context.Context, fn TransactionFn) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(d, tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("transaction failed: %w; rollback failed: %w", err, rollbackErr)
		}

		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (d *Database) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

func (d *Database) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

func (d *Database) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.db.QueryRow(query, args...)
}

func (d *Database) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}
