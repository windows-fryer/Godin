package main

import (
	"go.uber.org/zap"
	"wednesday.wtf/godin/internal/config"
	"wednesday.wtf/godin/internal/database"
	"wednesday.wtf/godin/internal/logger"
	"wednesday.wtf/godin/internal/server"
)

func createConfig() (*config.Config, error) {
	c, err := config.New()

	if err != nil {
		return nil, err
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}

func createLogger(config *config.Config) (*zap.Logger, error) {
	return logger.New(config.Development)
}

func main() {
	cfg, err := createConfig()

	if err != nil {
		panic(err)
	}

	log, err := createLogger(cfg)

	if err != nil {
		panic(err)
	}

	db, err := database.New(log, cfg)

	log.Info("Connection established to database")

	if err != nil {
		panic(err)
	}

	srv := server.New(log, db, cfg)

	log.Info("Connection established to server")

	if err := srv.Start(); err != nil {
		panic(err)
	}
}
