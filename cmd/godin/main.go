package main

import (
	"go.uber.org/zap"
	"wednesday.wtf/godin/internal/config"
	"wednesday.wtf/godin/internal/database"
	"wednesday.wtf/godin/internal/logger"
	"wednesday.wtf/godin/internal/server"
)

func createConfig() (*config.Config, error) {
	config, err := config.New()

	if err != nil {
		return nil, err
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func createLogger(config *config.Config) (*zap.Logger, error) {
	return logger.New(config.Development)
}

func main() {
	cfg, err := createConfig()

	if err != nil {
		panic(err)
	}

	logger, err := createLogger(cfg)

	if err != nil {
		panic(err)
	}

	db, err := database.New(logger, cfg)

	if err != nil {
		panic(err)
	}

	srv := server.New(logger, db, cfg)

	if err := srv.Start(); err != nil {
		panic(err)
	}
}
