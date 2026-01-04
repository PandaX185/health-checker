package database

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNew_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	logger := zap.NewNop()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://healthcheck:healthcheck@localhost:5432/healthcheck?sslmode=disable"
	}

	pool, err := New(ctx, dbURL, logger)

	if err != nil {
		// If database is not available, skip test
		t.Skip("Database not available:", err)
	}

	assert.NotNil(t, pool)
}

func TestNew_InvalidURL(t *testing.T) {

	ctx := context.Background()
	logger := zap.NewNop()

	// Test with a URL that will definitely fail to connect (non-existent host)
	// Use a short timeout to avoid waiting too long
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	pool, err := New(ctx, "postgres://user:pass@nonexistenthost:5432/db?sslmode=disable&connect_timeout=1", logger)
	assert.Error(t, err)
	assert.Nil(t, pool)
}

func TestNew_ParseConfigError(t *testing.T) {

	ctx := context.Background()
	logger := zap.NewNop()

	// Test with a completely invalid URL format
	pool, err := New(ctx, "invalid-url-format", logger)
	assert.Error(t, err)
	assert.Nil(t, pool)
}
