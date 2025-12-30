package monitor

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, service Service) error
	ListServices(ctx context.Context) ([]Service, error)
	ClaimDueServices(ctx context.Context) ([]Service, error)
	CreateHealthCheck(ctx context.Context, check HealthCheck) error
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, service Service) error {
	query := `
		INSERT INTO services (name, url, check_interval, next_run_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(ctx, query, service.Name, service.URL, service.CheckInterval, service.NextRunAt)
	return err
}

func (r *PostgresRepository) ListServices(ctx context.Context) ([]Service, error) {
	query := `
		SELECT id, name, url, check_interval, next_run_at, created_at
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
		err := rows.Scan(&service.ID, &service.Name, &service.URL, &service.CheckInterval, &service.NextRunAt, &service.CreatedAt)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	return services, nil
}

func (r *PostgresRepository) ClaimDueServices(ctx context.Context) ([]Service, error) {
	query := `
		update services 
		set next_run_at = now() + make_interval(secs => check_interval)
		where id in (
			select id from services 
			where next_run_at <= now()
			for update skip locked
		)
		returning id, name, url, check_interval, next_run_at, created_at
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []Service
	for rows.Next() {
		var service Service
		err := rows.Scan(&service.ID, &service.Name, &service.URL, &service.CheckInterval, &service.NextRunAt, &service.CreatedAt)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	return services, nil
}

func (r *PostgresRepository) CreateHealthCheck(ctx context.Context, check HealthCheck) error {
	query := `
		INSERT INTO health_checks (service_id, status)
		VALUES ($1, $2)
	`
	_, err := r.db.Exec(ctx, query, check.ServiceID, check.Status)
	return err
}
