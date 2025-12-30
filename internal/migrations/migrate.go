package migrations

import "github.com/jackc/pgx/v5/pgxpool"

var migrations = []func(*pgxpool.Pool) error{
	CreateServicesTable,
	CreateUsersTable,
	CreateHealthChecksTable,
}

var rollbacks = []func(*pgxpool.Pool) error{
	RollbackCreateServicesTable,
	RollbackCreateUsersTable,
	RollbackCreateHealthChecksTable,
}

func Migrate(db *pgxpool.Pool) error {
	var err error
	for _, m := range migrations {
		err = m(db)
		if err != nil {
			return err
		}
	}

	return nil
}

func Rollback(db *pgxpool.Pool) error {
	var err error
	for _, r := range rollbacks {
		err = r(db)
		if err != nil {
			return err
		}
	}

	return nil
}
