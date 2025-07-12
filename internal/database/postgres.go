package database

import (
	"database/sql"
	"go.uber.org/zap"
	"wednesday.wtf/godin/internal/config"

	_ "github.com/lib/pq"
)

type Database struct {
	log    *zap.Logger
	db     *sql.DB
	config *config.Config
}

func New(log *zap.Logger, cfg *config.Config) (*Database, error) {
	db, err := sql.Open("postgres", cfg.PostgresConnectionString)

	if err != nil {
		return nil, err
	}

	return &Database{
		log:    log,
		db:     db,
		config: cfg,
	}, nil
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
