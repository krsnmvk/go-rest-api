package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type Queries struct {
	DB DBTX
}

func NewQueries(db DBTX) *Queries {
	return &Queries{DB: db}
}

func (q *Queries) WithTx(tx pgx.Tx) *Queries {
	return &Queries{DB: tx}
}
