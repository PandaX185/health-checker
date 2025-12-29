package monitor

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, service Service) error
	ListServices(ctx context.Context) ([]Service, error)
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

func (r *PostgresRepository) ListServices(ctx context.Context) ([]Service, error) {
	query := `
		SELECT id, name, url, check_interval, created_at
		FROM services
		order by created_at desc
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []Service
	for rows.Next() {
		var service Service
		err := rows.Scan(&service.ID, &service.Name, &service.URL, &service.CheckInterval, &service.CreatedAt)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	return services, nil
}
