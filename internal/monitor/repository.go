package monitor

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, service Service) error
	ListServices(ctx context.Context) ([]Service, error)
	ClaimDueServices(ctx context.Context) ([]Service, error)
	CreateHealthCheck(ctx context.Context, check HealthCheck) error
	GetHealthChecksByServiceID(ctx context.Context, serviceID, page, limit int) ([]HealthCheck, error)
	GetLatestHealthCheck(ctx context.Context, serviceID int) (*HealthCheck, error)
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
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
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, query)
	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}
	defer rows.Close()
	defer tx.Commit(ctx)

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
		INSERT INTO health_checks (service_id, status, latency)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(ctx, query, check.ServiceID, check.Status, check.Latency)
	return err
}

func (r *PostgresRepository) GetHealthChecksByServiceID(ctx context.Context, serviceID, page, limit int) ([]HealthCheck, error) {
	query := `
		SELECT id, service_id, status, latency, created_at
		FROM health_checks
		WHERE service_id = $1
		ORDER BY created_at DESC
		OFFSET $2 LIMIT $3
	`
	rows, err := r.db.Query(ctx, query, serviceID, (page-1)*limit, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checks []HealthCheck
	for rows.Next() {
		var check HealthCheck
		err := rows.Scan(&check.ID, &check.ServiceID, &check.Status, &check.Latency, &check.CreatedAt)
		if err != nil {
			return nil, err
		}
		checks = append(checks, check)
	}

	return checks, nil
}

func (r *PostgresRepository) GetLatestHealthCheck(ctx context.Context, serviceID int) (*HealthCheck, error) {
	query := `
		SELECT id, service_id, status, latency, created_at
		FROM health_checks
		WHERE service_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	var check HealthCheck
	err := r.db.QueryRow(ctx, query, serviceID).Scan(&check.ID, &check.ServiceID, &check.Status, &check.Latency, &check.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &check, nil
}
