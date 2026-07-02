package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nagomi-project/database/internal/gen"
)

const (
	transactionTimeout = time.Second * 60
)

type Database struct {
	pool    *pgxpool.Pool
	queries *gen.Queries
}

// NewDatabase will return a new Database object that can be used for database operations.
func NewDatabase(pool *pgxpool.Pool) *Database {
	queries := gen.New(pool)

	return (&Database{
		pool:    pool,
		queries: queries,
	}).init()
}

// init will initalize all of the stores for the database.
func (db *Database) init() *Database {
	return db
}

// Queries will return all of the available queries.
func (db *Database) Queries() *gen.Queries {
	return db.queries
}

// WithTx will allow running a transaction inside of a callback function.
func (db *Database) WithTx(ctx context.Context, txFn func(db *Database) error) error {
	ctx, cancel := context.WithTimeout(ctx, transactionTimeout)
	defer cancel()

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	queries := db.queries.WithTx(tx)
	if err := txFn((&Database{
		pool:    db.pool,
		queries: queries,
	}).init()); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	committed = true
	return nil
}
