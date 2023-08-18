// Package solc provides utilities for managing and interacting with the Solidity compiler.
package solc

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// GetProductionLogger creates and returns a new production logger using the zap library.
// The production logger is optimized for performance and is suitable for use in production environments.
//
// Returns:
// - A pointer to the created zap.Logger instance.
// - An error if there's any issue creating the logger.
func GetProductionLogger() (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	return logger, err
}

// GetDevelopmentLogger creates and returns a new development logger using the zap library.
// The development logger is optimized for development and debugging, providing more detailed logs.
// The log level can be set using the provided level parameter.
//
// Parameters:
// - level: The desired log level (e.g., DebugLevel, InfoLevel, etc.).
//
// Returns:
// - A pointer to the created zap.Logger instance.
// - An error if there's any issue creating the logger.
func GetDevelopmentLogger(level zapcore.Level) (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(level)
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, err := config.Build()
	return logger, err
}
