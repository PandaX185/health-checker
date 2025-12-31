package auth

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, user User) error
	GetUserByUsername(ctx context.Context, username string) (User, error)
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, user User) error {
	query := `
		INSERT INTO users (username, password)
		VALUES ($1, $2)
	`

	_, err := r.db.Exec(ctx, query, user.Username, user.Password)
	return err
}

func (r *PostgresRepository) GetUserByUsername(ctx context.Context, username string) (User, error) {
	query := `
	select * from users where username=$1
	`
	row := r.db.QueryRow(ctx, query, username)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt)

	return user, err
}
