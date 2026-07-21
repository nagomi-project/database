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

	Modules   *modules
	Guild     *guild
	ActionLog *actionLog

	EventLog    *eventLog
	Infractions *infractions
	OAuth       *oAuth
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
	db.Modules = newModules(db)
	db.Guild = newGuildSettings(db)
	db.ActionLog = newActionLog(db)

	db.Infractions = newInfractions(db)
	db.EventLog = newEventLog(db)
	db.OAuth = newOAuth(db)

	return db
}

// withTx will allow running a transaction inside of a callback function.
// Transactions are given 1 minute to execute, otherwise it will timeout.
//
// This is useful in circumstances where you need to execute other logic inside of the transaction.
func (db *Database) withTx(ctx context.Context, txFn TransactionFunc) error {
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
