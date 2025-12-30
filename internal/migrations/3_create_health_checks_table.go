package migrations

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateHealthChecksTable(db *pgxpool.Pool) error {
	query := `
	CREATE TABLE IF NOT EXISTS health_checks (
		id SERIAL PRIMARY KEY,
		service_id INTEGER NOT NULL,
		status VARCHAR(50) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT clock_timestamp(),
		FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE
	);
	`

	_, err := db.Exec(context.Background(), query)
	return err
}

func RollbackCreateHealthChecksTable(db *pgxpool.Pool) error {
	query := `DROP TABLE IF EXISTS health_checks;`
	_, err := db.Exec(context.Background(), query)
	return err
}
