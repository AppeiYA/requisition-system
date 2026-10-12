package db

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func newMockDB(t *testing.T) (*DB, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	return New(sqlx.NewDb(sqlDB, "postgres")), mock
}

func TestTransactionManager_WithinTransaction_Commit(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectBegin()
	mock.ExpectCommit()

	m := NewTransactionManager(db)

	var gotExec Executor
	committed := false
	err := m.WithinTransaction(context.Background(), func(ctx context.Context) error {
		committed = true
		gotExec = db.Executor(ctx)
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if !committed {
		t.Fatal("expected fn to be called")
	}
	if _, ok := gotExec.(*sqlx.Tx); !ok {
		t.Errorf("expected transaction executor in context, got %T", gotExec)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestTransactionManager_WithinTransaction_RollsBackOnError(t *testing.T) {
	db, mock := newMockDB(t)
	fnErr := errors.New("fn failed")

	mock.ExpectBegin()
	mock.ExpectRollback()

	m := NewTransactionManager(db)

	err := m.WithinTransaction(context.Background(), func(ctx context.Context) error {
		return fnErr
	})
	if !errors.Is(err, fnErr) {
		t.Fatalf("expected fnErr, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestTransactionManager_WithinTransaction_CommitError(t *testing.T) {
	db, mock := newMockDB(t)
	commitErr := errors.New("commit failed")

	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(commitErr)
	// The deferred rollback short-circuits with sql.ErrTxDone after a failed
	// commit, so no driver-level rollback is expected.

	m := NewTransactionManager(db)

	err := m.WithinTransaction(context.Background(), func(ctx context.Context) error {
		return nil
	})
	if !errors.Is(err, commitErr) {
		t.Fatalf("expected commitErr, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDB_Ping(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectPing()

	if err := db.Ping(context.Background()); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDB_Executor_Default(t *testing.T) {
	db, _ := newMockDB(t)

	if got := db.Executor(context.Background()); got != db.conn {
		t.Errorf("expected default database executor, got %T", got)
	}
}

func TestDB_Executor_ContextOverride(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectBegin()

	tx, err := db.conn.BeginTxx(context.Background(), nil)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}

	if got := db.Executor(WithExecutor(context.Background(), tx)); got.(*sqlx.Tx) != tx {
		t.Errorf("expected transaction executor from context, got %T", got)
	}

	mock.ExpectRollback()
	if err := tx.Rollback(); err != nil {
		t.Fatalf("failed to roll back tx: %v", err)
	}
}
