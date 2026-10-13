package db

import (
	"context"
	"time"

	"github.com/AppeiYA/requisition-system/internal/shared/config"
	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
)

func Connect(cfg config.DBConfig) (*DB, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	conn, err := sqlx.ConnectContext(ctx, "postgres", cfg.DSN())
	if err != nil {
		return nil, err
	}

	conn.SetMaxOpenConns(cfg.MaxOpenConnections)
	conn.SetMaxIdleConns(cfg.MaxIdleConnections)
	conn.SetConnMaxLifetime(
		time.Duration(cfg.MaxLifetimeMinutes) * time.Minute,
	)

	db := New(conn)
	if err := db.Ping(ctx); err != nil {
		return nil, err
	}

	return db, nil
}