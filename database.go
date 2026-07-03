package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nagomi-project/database/internal/gen"
)

type TransactionFunc func(ctx context.Context, txDb *Database) error

const (
	transactionTimeout = time.Second * 60
)

type Database struct {
	pool    *pgxpool.Pool
	dbtx    gen.DBTX
	queries *gen.Queries
}

// NewDatabase will return a new Database object that can be used for database operations.
func NewDatabase(pool *pgxpool.Pool) *Database {
	queries := gen.New()

	return (&Database{
		pool:    pool,
		dbtx:    pool,
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

// Tx will return the current transaction.
//
// This has to be passed in whenever a query is called.
func (db *Database) Tx() gen.DBTX {
	return db.dbtx
}

// WithTx will allow running a transaction inside of a callback function.
// Transactions are given 1 minute to execute, otherwise it will timeout.
//
// This is useful in circumstances where you need to execute other logic inside of the transaction.
func (db *Database) WithTx(ctx context.Context, txFn TransactionFunc) error {
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

	// uses a new database object with the transaction
	if err := txFn(ctx, (&Database{
		pool:    db.pool,
		dbtx:    tx,
		queries: db.queries,
	}).init()); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	committed = true
	return nil
}
