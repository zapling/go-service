package business

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/zapling/go-service/internal/database"
)

// Client holds the business logic of the application.
type Client struct {
	db *pgxpool.Pool
}

// New creates a new business client.
func New(db *pgxpool.Pool) *Client {
	return &Client{db: db}
}

func (c *Client) CreateFoo(ctx context.Context, bar string) error {
	log := zerolog.Ctx(ctx)

	tx, err := c.db.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to start db transaction: %w", err)
	}

	defer func() {
		if err := tx.Rollback(context.Background()); err != nil && errors.Is(err, pgx.ErrTxClosed) {
			log.Err(err).Msg("Failed to rollback db transaction")
		}
	}()

	db := database.New(tx)
	err = db.Foo.Insert(ctx, database.Foo{Bar: bar})
	if err != nil {
		return fmt.Errorf("failed to insert foo: %w", err)
	}

	return tx.Commit(context.Background())
}
