package database

import (
	"context"

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
	return &Client{conn: conn}
}

// Client holds the database interaction logic.
// Add methods directly to it or add subclients to seperate logic further.
type Client struct {
	Foo fooClient

	conn dbOrTx // The db connection to use.
}

// dbOrTx is an interface that satisfies both a normal connection and a db transaction.
type dbOrTx interface {
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
