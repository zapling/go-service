package webservice

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func getDatabaseConn(ctx context.Context, connStr string) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get new db pool: %w", err)
	}

	if err := db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping to db failed: %w", err)
	}

	return db, nil
}
