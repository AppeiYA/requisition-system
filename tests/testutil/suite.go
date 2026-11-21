package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	appdb "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/jmoiron/sqlx"
)

type Suite struct {
	DB    *TestDatabase
	SQLX  *sqlx.DB
	AppDB *appdb.DB
	Redis *TestRedis
}

func NewSuite(t *testing.T, migrationsPath string) *Suite {
	t.Helper()

	resolvedMigrationsPath, err := ResolveMigrationsPath(migrationsPath)
	if err != nil {
		t.Fatalf("resolve migrations path: %v", err)
	}

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

	sqlDB, err := sqlx.Open("postgres", db.DSN)
	if err != nil {
		_ = db.Close()
		t.Fatalf("open test database: %v", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		_ = db.Close()
		t.Fatalf("ping test database connection: %v", err)
	}

	appDB := appdb.New(sqlDB)

	// Register cleanup immediately so migration failures
	// don't leave containers running.
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Errorf("cleanup test sqlx database: %v", err)
		}
		if err := db.Close(); err != nil {
			t.Errorf("cleanup test database: %v", err)
		}
	})

	// Apply application migrations.
	if err := RunMigrations(db.DSN, resolvedMigrationsPath); err != nil {
		_ = sqlDB.Close()
		t.Fatalf("run database migrations: %v", err)
	}

	return &Suite{
		DB:    db,
		SQLX:  sqlDB,
		AppDB: appDB,
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

func ResolveMigrationsPath(migrationsPath string) (string, error) {
	if migrationsPath == "" {
		return "", fmt.Errorf("migrations path is empty")
	}

	candidates := []string{migrationsPath}
	if !filepath.IsAbs(migrationsPath) {
		workingDirectory, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("get working directory: %w", err)
		}

		for directory := workingDirectory; ; directory = filepath.Dir(directory) {
			candidates = append(
				candidates,
				filepath.Join(directory, migrationsPath),
			)

			parent := filepath.Dir(directory)
			if parent == directory {
				break
			}
		}
	}

	for _, candidate := range candidates {
		absolutePath, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if info, err := os.Stat(absolutePath); err == nil && info.IsDir() {
			return absolutePath, nil
		}
	}

	return "", fmt.Errorf("migrations directory not found: %s", migrationsPath)
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
