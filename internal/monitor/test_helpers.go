package monitor

import (
	"context"
	"health-checker/internal/migrations"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SetupTestDatabase creates tables in the test database
func SetupTestDatabase(t *testing.T) {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:root@localhost:5432/health_checker?sslmode=disable"
	}

	// Connect to test database
	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := migrations.Migrate(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}
}
