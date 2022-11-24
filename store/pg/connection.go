package pg

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/leophys/userz/store/pg/postgres"
)

type preparer interface {
	Prepare(ctx context.Context, name, statement string) (*pgconn.StatementDescription, error)
}

// db represents a connection which is not in transaction.
type db interface {
	Begin(context.Context) (tx, error)
	postgres.DBTX
	preparer
}

// tx represents a connection with an open transaction.
type tx interface {
	Commit(context.Context) error
	Rollback(context.Context) error
	postgres.DBTX
	preparer
}

var _ db = &PGPooledConn{}

// PGPooledConn wraps a github.com/jackc/pgx/v4/pgxpool.Pool and implements
// the db interface.
type PGPooledConn struct {
	pool *pgxpool.Pool
}

func (c *PGPooledConn) Begin(ctx context.Context) (tx, error) {
	return c.pool.Begin(ctx)
}

func (c *PGPooledConn) Exec(ctx context.Context, statement string, params ...interface{}) (pgconn.CommandTag, error) {
	return c.pool.Exec(ctx, statement, params...)
}

func (c *PGPooledConn) Query(ctx context.Context, statement string, params ...interface{}) (pgx.Rows, error) {
	return c.pool.Query(ctx, statement, params...)
}

func (c *PGPooledConn) QueryRow(ctx context.Context, statement string, params ...interface{}) pgx.Row {
	return c.pool.QueryRow(ctx, statement, params...)
}

func (c *PGPooledConn) Prepare(ctx context.Context, name, statement string) (*pgconn.StatementDescription, error) {
	conn, err := c.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	return conn.Conn().Prepare(ctx, name, statement)
}
