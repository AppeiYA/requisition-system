package integration_test

import (
	"context"
	"testing"

	"github.com/AppeiYA/requisition-system/tests/testutil"
)

func TestPostgresConnection(t *testing.T) {
	suite := testutil.NewSuite(
		t,
		"../../migrations",
	)

	var result int

	err := suite.DB.Pool.QueryRow(
		context.Background(),
		"SELECT 1",
	).Scan(&result)

	if err != nil {
		t.Fatalf("execute query: %v", err)
	}

	if result != 1 {
		t.Fatalf("expected 1, got %d", result)
	}
}

func TestPostgresMigrationsAndReset(t *testing.T) {
	suite := testutil.NewSuite(t, "../../migrations")
	ctx := context.Background()

	var usersTableCount int
	err := suite.DB.Pool.QueryRow(
		ctx,
		`SELECT COUNT(*) FROM information_schema.tables
		 WHERE table_schema = 'public' AND table_name = 'users'`,
	).Scan(&usersTableCount)
	if err != nil {
		t.Fatalf("check users table: %v", err)
	}
	if usersTableCount != 1 {
		t.Fatalf("expected users table to exist, got %d matches", usersTableCount)
	}

	_, err = suite.DB.Pool.Exec(ctx, `
		INSERT INTO users (
			id, first_name, last_name, email, password_hash
		) VALUES (
			'integration-reset-user',
			'Integration',
			'Reset',
			'integration-reset@example.com',
			'not-a-real-password-hash'
		)
	`)
	if err != nil {
		t.Fatalf("seed users table: %v", err)
	}

	if err := suite.ResetDatabase(ctx); err != nil {
		t.Fatalf("reset database: %v", err)
	}

	var usersCount int
	err = suite.DB.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&usersCount)
	if err != nil {
		t.Fatalf("count users after reset: %v", err)
	}
	if usersCount != 0 {
		t.Fatalf("expected users table to be empty, got %d rows", usersCount)
	}
}
