package database

import (
	"database/sql"
	"go.uber.org/zap"
	"time"
	"wednesday.wtf/godin/internal/config"

	_ "github.com/lib/pq"
)

type Database struct {
	log    *zap.Logger
	db     *sql.DB
	config *config.Config
}

func (d *Database) clearExpiredSessions() error {
	if _, err := d.Transaction(func(d *Database, tx *sql.Tx) (*sql.Result, error) {
		rows, err := tx.Exec(`DELETE FROM godin.eris.sessions WHERE expiration_time < NOW()`)

		if err != nil {
			return nil, err
		}

		rowsAffected, err := rows.RowsAffected()

		if err != nil {
			return nil, err
		}

		d.log.Debug("Cleaned expired sessions", zap.Int("rows_affected", int(rowsAffected)))

		return nil, nil
	}); err != nil {
		return err
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

	go func() {
		ticker := time.NewTicker(1 * time.Minute)

		defer ticker.Stop()

		for range ticker.C {
			if err := database.clearExpiredSessions(); err != nil {
				panic(err)
			}
		}
	}()

	return &database, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

type TransactionFn func(d *Database, tx *sql.Tx) (*sql.Result, error)

func (d *Database) Transaction(fn TransactionFn) (*sql.Result, error) {
	tx, err := d.db.Begin()

	if err != nil {
		return nil, err
	}

	result, err := fn(d, tx)

	if err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, err
		}

		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return result, nil
}

func (d *Database) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

func (d *Database) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.db.QueryRow(query, args...)
}
