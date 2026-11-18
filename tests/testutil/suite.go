package testutil

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

type Suite struct {
	DB    *TestDatabase
	Redis *TestRedis
}

func NewSuite(t *testing.T, migrationsPath string) *Suite {
	t.Helper()

	// Give infrastructure setup enough time to complete.
	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Minute,
	)
	defer cancel()

	// Start PostgreSQL.
	db, err := SetupTestDatabase(ctx)
	if err != nil {
		t.Fatalf("setup test database: %v", err)
	}

	// Register cleanup immediately so migration failures
	// don't leave containers running.
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("cleanup test database: %v", err)
		}
	})

	// Apply application migrations.
	if err := RunMigrations(db.DSN, migrationsPath); err != nil {
		t.Fatalf("run database migrations: %v", err)
	}

	return &Suite{
		DB: db,
	}
}

func (s *Suite) SetupRedis(t *testing.T) *TestRedis {
	t.Helper()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Minute,
	)
	defer cancel()

	r, err := SetupTestRedis(ctx)
	if err != nil {
		t.Fatalf("setup test redis: %v", err)
	}

	t.Cleanup(func() {
		if err := r.Close(); err != nil {
			t.Errorf("cleanup test redis: %v", err)
		}
	})

	s.Redis = r

	return r
}

// ResetDatabase removes application and migration data while preserving the
// migrated schema.
func (s *Suite) ResetDatabase(ctx context.Context) error {
	if s == nil || s.DB == nil || s.DB.Pool == nil {
		return fmt.Errorf("test database is not initialized")
	}

	rows, err := s.DB.Pool.Query(ctx, `
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = 'public'
		  AND tablename NOT LIKE 'pg_%'
		  AND tablename <> 'schema_migrations'
	`)
	if err != nil {
		return fmt.Errorf("list test database tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return fmt.Errorf("scan test database table: %w", err)
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate test database tables: %w", err)
	}

	if len(tables) == 0 {
		return nil
	}

	quoted := make([]string, 0, len(tables))
	for _, table := range tables {
		quoted = append(quoted, fmt.Sprintf("%q", table))
	}

	_, err = s.DB.Pool.Exec(
		ctx,
		"TRUNCATE TABLE "+strings.Join(quoted, ", ")+" RESTART IDENTITY CASCADE",
	)
	if err != nil {
		return fmt.Errorf("reset test database: %w", err)
	}

	return nil
}

// ResetRedis removes all keys from the suite's Redis database.
func (s *Suite) ResetRedis(ctx context.Context) error {
	if s == nil || s.Redis == nil || s.Redis.Client == nil {
		return fmt.Errorf("test redis is not initialized")
	}

	if err := s.Redis.Client.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("reset test redis: %w", err)
	}

	return nil
}

func (s *Suite) Reset(ctx context.Context) error {
	if err := s.ResetDatabase(ctx); err != nil {
		return err
	}
	if s.Redis != nil {
		if err := s.ResetRedis(ctx); err != nil {
			return err
		}
	}
	return nil
}
