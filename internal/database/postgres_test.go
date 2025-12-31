package database

import (
	"context"
	"os"
	"testing"

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
	// Don't close pool as it uses singleton pattern
}

func TestNew_InvalidURL(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	_, err := New(ctx, "invalid-connection-string", logger)
	assert.Error(t, err)
}
