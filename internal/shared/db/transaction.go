package db

import (
	"context"
	"database/sql"
	"errors"
	"log"
)

type TransactionManager struct {
	db *DB
}

func NewTransactionManager(database *DB) *TransactionManager {
	return &TransactionManager{
		db: database,
	}
}

func (m *TransactionManager) WithinTransaction(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	tx, err := m.db.conn.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("failed to roll back transaction: %v", err)
		}
	}()

	ctx = WithExecutor(ctx, tx)

	if err := fn(ctx); err != nil {
		return err
	}

	return tx.Commit()
}