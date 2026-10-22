package db

import (
	"context"
	"time"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"

	_ "github.com/lib/pq"
)

func Connect(cfg config.DBConfig) (*DB, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	// otelsql does not name the database system on its own since it dropped
	// the deprecated semconv: without this the spans and the
	// db.client.operation.duration measurements carry no db.system.name at
	// all, and neither a trace search nor a dashboard panel can tell them
	// apart from the ones of another database.
	conn, err := otelsql.Open(
		"postgres",
		cfg.DSN(),
		otelsql.WithAttributes(semconv.DBSystemNamePostgreSQL),
	)
	if err != nil {
		return nil, err
	}
	connx := sqlx.NewDb(conn, "postgres")

	connx.SetMaxOpenConns(cfg.MaxOpenConnections)
	connx.SetMaxIdleConns(cfg.MaxIdleConnections)
	connx.SetConnMaxLifetime(
		time.Duration(cfg.MaxLifetimeMinutes) * time.Minute,
	)

	db := New(connx)
	if err := db.Ping(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
