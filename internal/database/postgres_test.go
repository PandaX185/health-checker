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
	Reset()
	defer Reset()

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
	assert.Equal(t, pool, Get())
	// Don't close pool as it uses singleton pattern
}

func TestNew_InvalidURL(t *testing.T) {
	Reset()       // Reset singleton to ensure we test initialization logic
	defer Reset() // Clean up after test

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
	Reset() // Reset singleton
	defer Reset()

	ctx := context.Background()
	logger := zap.NewNop()

	// Test with a completely invalid URL format
	pool, err := New(ctx, "invalid-url-format", logger)
	assert.Error(t, err)
	assert.Nil(t, pool)
}

func TestGet(t *testing.T) {
	Reset()
	// Test Get when pool is nil
	pool := Get()
	assert.Nil(t, pool)
}

func TestClose(t *testing.T) {
	// Test Close function exists and can be called
	// This is safe to call multiple times
	Close()
}

func TestNew_Singleton(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	logger := zap.NewNop()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://healthcheck:healthcheck@localhost:5432/healthcheck?sslmode=disable"
	}

	// Call New multiple times
	pool1, err1 := New(ctx, dbURL, logger)
	pool2, err2 := New(ctx, dbURL, logger)

	// Both should succeed and return the same pool (singleton)
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotNil(t, pool1)
	assert.NotNil(t, pool2)
	assert.Equal(t, pool1, pool2, "New should return the same pool instance (singleton)")
}
