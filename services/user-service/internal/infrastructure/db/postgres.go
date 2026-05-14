package db

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

func ConnectPostgres(ctx context.Context, databaseURL string) (*sql.DB, error) {
	database, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	if err := database.PingContext(ctx); err != nil {
		return nil, err
	}
	return database, nil
}
