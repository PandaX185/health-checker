package migrations

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateServicesTable(db *pgxpool.Pool) error {
	query := `
	CREATE TABLE IF NOT EXISTS services (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		url VARCHAR(255) NOT NULL,
		check_interval INT NOT NULL,
		next_run_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT clock_timestamp(),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT clock_timestamp()
	);
	`

	_, err := db.Exec(context.Background(), query)
	return err
}

func RollbackCreateServicesTable(db *pgxpool.Pool) error {
	query := `DROP TABLE IF EXISTS services;`
	_, err := db.Exec(context.Background(), query)
	return err
}
