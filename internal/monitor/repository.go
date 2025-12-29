package monitor

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, service Service) error
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, service Service) error {
	query := `
		INSERT INTO services (name, url, check_interval)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(ctx, query, service.Name, service.URL, service.CheckInterval)
	return err
}
