package database

import (
	"database/sql"
	"os"

	"github.com/golang/glog"
	_ "github.com/lib/pq"
)

var PostgresClient *sql.DB

func Start() {
	connection := os.Getenv("POSTGRES_CONNECTION")

	db, err := sql.Open("postgres", connection)

	if err != nil {
		glog.Fatalf("Failed to open database connection: %v", err)
	}

	if err = db.Ping(); err != nil {
		glog.Fatalf("Failed to ping database: %v", err)
	}

	PostgresClient = db
}

func Stop() {
	if PostgresClient != nil {
		PostgresClient.Close()
	}

	glog.Info("Database connection closed")
}
