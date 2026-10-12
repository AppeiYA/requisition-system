package db

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type DB struct {
	conn *sqlx.DB
}

func New(conn *sqlx.DB) *DB {
	return &DB{
		conn: conn,
	}
}

func (db *DB) Ping(ctx context.Context) error {
	return db.conn.PingContext(ctx)
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) Conn() *sqlx.DB {
	return db.conn
}