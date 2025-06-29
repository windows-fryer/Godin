package main

import (
	"go.uber.org/zap"
	"wednesday.wtf/godin/internal/config"
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
	config, err := createConfig()

	if err != nil {
		panic(err)
	}

	logger, err := createLogger(config)

	if err != nil {
		panic(err)
	}

	srv := server.New(logger, config)

	if err := srv.Start(); err != nil {
		panic(err)
	}
}
