package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a new zap logger based on the environment
func NewLogger() (*zap.Logger, error) {
	env := os.Getenv("ENV")
	logLevel := os.Getenv("LOG_LEVEL")

	var logger *zap.Logger
	var err error

	if env == "production" {
		// Production logger configuration
		config := zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		// Set log level for production
		if logLevel != "" {
			level, parseErr := zapcore.ParseLevel(strings.ToLower(logLevel))
			if parseErr == nil {
				config.Level.SetLevel(level)
			}
		}

		logger, err = config.Build()
	} else {
		// Development logger configuration
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

		// Set log level for development (default to debug)
		if logLevel != "" {
			level, parseErr := zapcore.ParseLevel(strings.ToLower(logLevel))
			if parseErr == nil {
				config.Level.SetLevel(level)
			}
		} else {
			// Default to debug in development
			config.Level.SetLevel(zapcore.DebugLevel)
		}

		logger, err = config.Build()
	}

	if err != nil {
		return nil, err
	}

	return logger, nil
}

// NewSugaredLogger creates a new sugared zap logger
func NewSugaredLogger() (*zap.SugaredLogger, error) {
	logger, err := NewLogger()
	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}
