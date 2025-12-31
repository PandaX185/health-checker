package auth

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()

	// Setup database connection
	dbURL := "postgres://healthcheck:healthcheck@localhost:5432/healthcheck?sslmode=disable"
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)
	defer pool.Close()

	repo := NewRepository(pool)

	t.Run("Create", func(t *testing.T) {
		user := User{
			Username:  "testuser",
			Password:  "hashedpassword",
			CreatedAt: time.Now(),
		}

		err := repo.Create(ctx, user)
		assert.NoError(t, err)
	})

	t.Run("GetUserByUsername", func(t *testing.T) {
		user, err := repo.GetUserByUsername(ctx, "testuser")
		assert.NoError(t, err)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "hashedpassword", user.Password)
		assert.NotZero(t, user.ID)
	})

	t.Run("GetUserByUsername_NotFound", func(t *testing.T) {
		_, err := repo.GetUserByUsername(ctx, "nonexistent")
		assert.Error(t, err)
	})
}
