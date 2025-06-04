package main

import (
	"github.com/godin/internal/database"
	"github.com/godin/internal/server"
	"github.com/joho/godotenv"
)

func initializeDependencies() {
	database.Start()
	server.Start()
}

func main() {
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}

	initializeDependencies()

	select {}
}
