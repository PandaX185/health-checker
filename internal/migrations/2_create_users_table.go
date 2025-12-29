package migrations

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateUsersTable(db *pgxpool.Pool) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT clock_timestamp()
	);
	`

	_, err := db.Exec(context.Background(), query)
	return err
}

func RollbackCreateUsersTable(db *pgxpool.Pool) error {
	query := `DROP TABLE IF EXISTS users;`
	_, err := db.Exec(context.Background(), query)
	return err
}
