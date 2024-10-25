package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// New create a new database client with the provided conn.
// Conn can either be a normal db connection or a db transaction.
func New(conn dbOrTx) *Client {
	if conn == nil {
		panic("Tried to create a new database client with a nil conn")
	}
	c := &Client{conn: conn}
	c.Foo = fooClient{client: c}
	return c
}

// Client holds the database interaction logic.
// Add methods directly to it or add subclients to seperate logic further.
type Client struct {
	Foo fooClient

	conn dbOrTx // The db connection to use.
}

type Conn interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	dbOrTx
}

// dbOrTx is an interface that satisfies both a normal connection and a db transaction.
type dbOrTx interface {
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// TransactionBlock runs the provided function within a database transaction.
// If the returnErr is nil it will commit, or rollback the db transaction if there was an error.
func TransactionBlock(conn Conn, fn func(tx pgx.Tx) error) (returnErr error) {
	tx, err := conn.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to start db transaction: %w", err)
	}

	defer func() {
		errRollback := tx.Rollback(context.Background())
		if errRollback != nil && !errors.Is(errRollback, pgx.ErrTxClosed) {
			returnErr = fmt.Errorf("%w: %v", errRollback, returnErr)
		}
	}()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(context.Background()); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		return fmt.Errorf("failed to commit db transaction: %w", err)
	}

	return nil
}
