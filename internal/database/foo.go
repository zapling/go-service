package database

import (
	"context"
)

type fooClient struct {
	client *Client
}

type Foo struct {
	Bar string
}

func (f *fooClient) Insert(ctx context.Context, foo Foo) error {
	query := `
	INSERT INTO foo (bar) VALUES ($1)
	`
	_, err := f.client.conn.Exec(ctx, query, foo.Bar)
	return err
}

func (f *fooClient) Get(ctx context.Context, bar string) (*Foo, error) {
	query := `
	SELECT bar FROM foo WHERE bar = $1
	`
	var foo Foo
	err := f.client.conn.QueryRow(ctx, query, bar).Scan(
		&foo.Bar,
	)
	if err != nil {
		return nil, err
	}
	return &foo, nil
}
