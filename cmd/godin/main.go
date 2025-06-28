package main

import (
	"wednesday.wtf/godin/internal/logger"
)

func main() {
	zapLogger, err := logger.New(true)

	if err != nil {
		panic(err)
	}

	zapLogger.Info("Hello, World!")
}
