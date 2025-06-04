package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/godin/internal/database"
	"github.com/godin/internal/discord"
	"github.com/godin/internal/server"
	"github.com/joho/godotenv"
)

func initializeDependencies() {
	database.Start()
	server.Start()
	discord.Start()
}

func shutdownDependencies() {
	database.Stop()
	discord.Stop()
	server.Stop()
}

func main() {
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}

	initializeDependencies()

	defer shutdownDependencies()

	termChan := make(chan os.Signal, 1)

	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	<-termChan
}
