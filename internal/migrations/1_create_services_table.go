package migrations

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func CreateServicesTable(conn *pgx.Conn) error {
	query := `
	CREATE TABLE IF NOT EXISTS services (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		url VARCHAR(255) NOT NULL,
		check_interval INT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := conn.Exec(context.Background(), query)
	return err
}

func RollbackCreateServicesTable(conn *pgx.Conn) error {
	query := `DROP TABLE IF EXISTS services;`
	_, err := conn.Exec(context.Background(), query)
	return err
}
