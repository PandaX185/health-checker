package config

import (
	"context"

	"github.com/jackc/pgx/v5"
)

var dbConn *pgx.Conn

func NewDatabase(connStr string) *pgx.Conn {
	once.Do(func() {
		conn, err := pgx.Connect(context.Background(), connStr)
		if err != nil {
			panic(err)
		}
		dbConn = conn
	})

	return dbConn
}
