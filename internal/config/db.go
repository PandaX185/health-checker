package config

import (
	"context"
	"health-checker/internal/migrations"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

var dbConn *pgx.Conn
var dbOnce = &sync.Once{}

func GetDatabase() *pgx.Conn {
	return dbConn
}

func NewDatabase(connStr string) *pgx.Conn {
	dbOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		conn, err := pgx.Connect(ctx, connStr)
		if err != nil {
			logger.Fatal("Failed to connect to database", zap.Error(err))
		}

		if err := conn.Ping(ctx); err != nil {
			logger.Fatal("Failed to ping database", zap.Error(err))
		}

		dbConn = conn
	})

	return dbConn
}

func Migrate() error {
	var err error

	if err = migrations.CreateServicesTable(dbConn); err != nil {
		logger.Error("Failed to migrate database", zap.Error(err))
		return err
	}

	logger.Info("Database migration completed successfully")
	return nil
}

func Rollback() error {
	var err error

	if err = migrations.RollbackCreateServicesTable(dbConn); err != nil {
		logger.Error("Failed to rollback database", zap.Error(err))
		return err
	}

	logger.Info("Database rollback completed successfully")
	return nil
}
