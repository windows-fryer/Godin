package database

import (
	"database/sql"
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

func (d *Database) clearExpiredSessions() error {
	if _, err := d.Transaction(func(d *Database, tx *sql.Tx) (*sql.Result, error) {
		rows, err := tx.Query(`
			SELECT CONCAT('godin.', schema_name, '.sessions') AS full_table_name
			FROM information_schema.schemata 
			WHERE schema_name LIKE '%'
			AND EXISTS (
				SELECT 1 
				FROM information_schema.tables 
				WHERE table_schema = schemata.schema_name 
				AND table_name = 'sessions'
			)
		`)

		if err != nil {
			return nil, err
		}

		defer func(rows *sql.Rows) {
			err := rows.Close()

			if err != nil {
				panic(err)
			}
		}(rows)

		for rows.Next() {
			var schema string

			if err := rows.Scan(&schema); err != nil {
				return nil, err
			}

			if _, err := tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE expiration_time < NOW()", schema)); err != nil {
				return nil, err
			}
		}

		if err := rows.Err(); err != nil {
			return nil, err
		}

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

	if err := database.clearExpiredSessions(); err != nil {
		panic(err)
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
