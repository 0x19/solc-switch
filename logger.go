// Package solc provides utilities for managing and interacting with the Solidity compiler.
package solc

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// GetProductionLogger creates and returns a new production logger using the zap library.
// The production logger is optimized for performance and is suitable for use in production environments.
// The log level can be set using the provided level parameter.
func GetProductionLogger(level zapcore.Level) (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(level)
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, err := config.Build()
	return logger, err
}

// GetDevelopmentLogger creates and returns a new development logger using the zap library.
// The development logger is optimized for development and debugging, providing more detailed logs.
// The log level can be set using the provided level parameter.
func GetDevelopmentLogger(level zapcore.Level) (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(level)
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, err := config.Build()
	return logger, err
}
