package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(development bool) (*zap.Logger, error) {
	var config zap.Config

	if development {
		config = zap.NewDevelopmentConfig()

		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		config.OutputPaths = []string{"stdout"}
	} else {
		config = zap.NewProductionConfig()

		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	return config.Build(zap.AddStacktrace(zapcore.ErrorLevel))
}
