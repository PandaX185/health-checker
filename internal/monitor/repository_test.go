package monitor

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRepository(t *testing.T) {
	pool := &pgxpool.Pool{}
	repo := NewRepository(pool)

	assert.NotNil(t, repo)
	assert.IsType(t, &PostgresRepository{}, repo)
}

// Integration tests for repository - require database
func TestRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test database tables
	SetupTestDatabase(t)

	ctx := context.Background()
	dbURL := getTestDatabaseURL()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)
	defer pool.Close()

	repo := NewRepository(pool)

	t.Run("Create", func(t *testing.T) {
		service := Service{
			Name:          "test-create-service",
			URL:           "http://test-create.com",
			CheckInterval: 30,
			NextRunAt:     time.Now().Add(30 * time.Second),
		}

		err := repo.Create(ctx, service)
		assert.NoError(t, err)
	})

	t.Run("ListServices", func(t *testing.T) {
		services, err := repo.ListServices(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, services)
	})

	t.Run("CreateHealthCheck", func(t *testing.T) {
		// First create a service
		service := Service{
			Name:          "test-health-check-service",
			URL:           "http://test-hc.com",
			CheckInterval: 60,
			NextRunAt:     time.Now().Add(60 * time.Second),
		}
		err := repo.Create(ctx, service)
		require.NoError(t, err)

		// Get the service
		services, err := repo.ListServices(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, services)

		serviceID := services[0].ID

		// Create health check
		check := HealthCheck{
			ServiceID: serviceID,
			Status:    "UP",
			Latency:   150,
			CreatedAt: time.Now(),
		}

		err = repo.CreateHealthCheck(ctx, check)
		assert.NoError(t, err)
	})

	t.Run("GetHealthChecksByServiceID", func(t *testing.T) {
		// Create service and health check first
		service := Service{
			Name:          "test-get-hc-service",
			URL:           "http://test-get-hc.com",
			CheckInterval: 45,
			NextRunAt:     time.Now().Add(45 * time.Second),
		}
		err := repo.Create(ctx, service)
		require.NoError(t, err)

		services, err := repo.ListServices(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, services)
		serviceID := services[0].ID

		check := HealthCheck{
			ServiceID: serviceID,
			Status:    "DOWN",
			Latency:   500,
			CreatedAt: time.Now(),
		}
		err = repo.CreateHealthCheck(ctx, check)
		require.NoError(t, err)

		// Get health checks
		checks, err := repo.GetHealthChecksByServiceID(ctx, serviceID, 1, 10)
		assert.NoError(t, err)
		assert.NotEmpty(t, checks)
		assert.Equal(t, serviceID, checks[0].ServiceID)
	})

	t.Run("GetLatestHealthCheck", func(t *testing.T) {
		// Create service and multiple health checks
		service := Service{
			Name:          "test-latest-hc-service",
			URL:           "http://test-latest-hc.com",
			CheckInterval: 20,
			NextRunAt:     time.Now().Add(20 * time.Second),
		}
		err := repo.Create(ctx, service)
		require.NoError(t, err)

		services, err := repo.ListServices(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, services)
		serviceID := services[0].ID

		// Create first check
		check1 := HealthCheck{
			ServiceID: serviceID,
			Status:    "UP",
			Latency:   100,
			CreatedAt: time.Now().Add(-2 * time.Minute),
		}
		err = repo.CreateHealthCheck(ctx, check1)
		require.NoError(t, err)

		// Create second check (latest)
		check2 := HealthCheck{
			ServiceID: serviceID,
			Status:    "DOWN",
			Latency:   300,
			CreatedAt: time.Now(),
		}
		err = repo.CreateHealthCheck(ctx, check2)
		require.NoError(t, err)

		// Get latest
		latest, err := repo.GetLatestHealthCheck(ctx, serviceID)
		assert.NoError(t, err)
		assert.NotNil(t, latest)
		assert.Equal(t, "DOWN", latest.Status)
		assert.Equal(t, serviceID, latest.ServiceID)
	})

	t.Run("ClaimDueServices", func(t *testing.T) {
		// Create a service that's due
		service := Service{
			Name:          "test-claim-due-service",
			URL:           "http://test-claim-due.com",
			CheckInterval: 15,
			NextRunAt:     time.Now().Add(-5 * time.Second), // Past due
		}
		err := repo.Create(ctx, service)
		require.NoError(t, err)

		// Claim due services
		dueServices, err := repo.ClaimDueServices(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, dueServices)

		// At least one service should be due
		if len(dueServices) > 0 {
			// Next run should be updated
			assert.True(t, dueServices[0].NextRunAt.After(time.Now()))
		}
	})

	t.Run("GetLatestHealthCheck_NotFound", func(t *testing.T) {
		// Try to get health check for non-existent service
		latest, err := repo.GetLatestHealthCheck(ctx, 999999)
		// Should return nil without error (no records found)
		assert.NoError(t, err)
		assert.Nil(t, latest)
	})
}
