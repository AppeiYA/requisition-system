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
