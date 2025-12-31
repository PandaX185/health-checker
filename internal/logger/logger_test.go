package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNew(t *testing.T) {
	logger := New("test-env")
	assert.NotNil(t, logger)
}

func TestNew_DifferentEnvironments(t *testing.T) {
	prodLogger := New("production")
	devLogger := New("development")
	testLogger := New("test")

	assert.NotNil(t, prodLogger)
	assert.NotNil(t, devLogger)
	assert.NotNil(t, testLogger)
}

func TestLogger_Logging(t *testing.T) {
	logger := New("development")

	// These should not panic
	logger.Info("info message", zap.String("key", "value"))
	logger.Error("error message", zap.String("key", "value"))
	logger.Debug("debug message", zap.String("key", "value"))
}

func TestLogger_WithFields(t *testing.T) {
	logger := New("development")

	// Test logging with multiple fields
	logger.Info("user action",
		zap.String("user_id", "123"),
		zap.String("action", "login"),
		zap.Int("status_code", 200),
	)

	assert.NotNil(t, logger)
}

func TestNew_InvalidEnvironment(t *testing.T) {
	// Should not panic with invalid environment
	logger := New("invalid-env")
	assert.NotNil(t, logger)
	logger.Info("test message")
}
