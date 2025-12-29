package migrations

import "github.com/jackc/pgx/v5/pgxpool"

func Migrate(db *pgxpool.Pool) error {
	var err error
	err = CreateServicesTable(db)
	if err != nil {
		return err
	}

	return nil
}

func Rollback(db *pgxpool.Pool) error {
	var err error
	err = RollbackCreateServicesTable(db)
	if err != nil {
		return err
	}

	return nil
}
