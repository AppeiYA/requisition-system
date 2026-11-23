package audit_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/audit"
	appdb "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/generator"
	"github.com/AppeiYA/requisition-system/tests/testutil"
	"github.com/stretchr/testify/require"
)

func setupRecorder(t *testing.T) (*testutil.Suite, *audit.PostgresRecorder) {
	t.Helper()

	suite := testutil.NewSuite(t, "../../../../migrations")
	ctx := context.Background()

	_, err := suite.SQLX.ExecContext(ctx, `
		INSERT INTO users (id, first_name, last_name, email, password_hash, status, role)
		VALUES ('usr-actor', 'Ada', 'Admin', 'ada@example.com', 'hash', 'active', 'admin')
	`)
	require.NoError(t, err)

	recorder := audit.NewPostgresRecorder(
		appdb.NewRepository(suite.AppDB),
		&generator.IDGenerator{},
	)

	return suite, recorder
}

func TestPostgresRecorder_RecordsEntryWithRequestMetadata(t *testing.T) {
	suite, recorder := setupRecorder(t)

	ctx := audit.WithRequestMeta(context.Background(), audit.RequestMeta{
		RequestID: "req-1",
		IPAddress: "203.0.113.7",
		UserAgent: "test-agent",
	})

	err := recorder.Record(ctx, audit.Entry{
		ID:         "aud-test-1",
		ActorID:    "usr-actor",
		Action:     "user.status.update",
		TargetType: "user",
		TargetID:   "usr-target",
		Before:     map[string]any{"status": "active"},
		After:      map[string]any{"status": "suspended"},
		Reason:     "policy",
	})
	require.NoError(t, err)

	var row struct {
		ID         string  `db:"id"`
		ActorID    string  `db:"actor_id"`
		Action     string  `db:"action"`
		TargetType string  `db:"target_type"`
		TargetID   string  `db:"target_id"`
		Before     string  `db:"before_state"`
		After      string  `db:"after_state"`
		Reason     *string `db:"reason"`
		RequestID  *string `db:"request_id"`
		IPAddress  *string `db:"ip_address"`
		UserAgent  *string `db:"user_agent"`
	}

	err = suite.SQLX.GetContext(ctx, &row, `
		SELECT id, actor_id, action, target_type, target_id,
		       before_state::text AS before_state,
		       after_state::text AS after_state,
		       reason, request_id, host(ip_address) AS ip_address, user_agent
		FROM admin_audit_log
		WHERE id = $1
	`, "aud-test-1")
	require.NoError(t, err)

	require.Equal(t, "usr-actor", row.ActorID)
	require.Equal(t, "user.status.update", row.Action)
	require.Equal(t, "user", row.TargetType)
	require.Equal(t, "usr-target", row.TargetID)
	require.JSONEq(t, `{"status":"active"}`, row.Before)
	require.JSONEq(t, `{"status":"suspended"}`, row.After)
	require.NotNil(t, row.Reason)
	require.Equal(t, "policy", *row.Reason)
	require.NotNil(t, row.RequestID)
	require.Equal(t, "req-1", *row.RequestID)
	require.NotNil(t, row.IPAddress)
	require.Equal(t, "203.0.113.7", *row.IPAddress)
	require.NotNil(t, row.UserAgent)
	require.Equal(t, "test-agent", *row.UserAgent)
}

func TestPostgresRecorder_JoinsCallerTransaction(t *testing.T) {
	suite, recorder := setupRecorder(t)
	tx := appdb.NewTransactionManager(suite.AppDB)
	ctx := context.Background()

	rolledBack := audit.Entry{
		ID:         "aud-tx-rollback",
		ActorID:    "usr-actor",
		Action:     "user.role.update",
		TargetType: "user",
		TargetID:   "usr-target",
	}

	err := tx.WithinTransaction(ctx, func(ctx context.Context) error {
		require.NoError(t, recorder.Record(ctx, rolledBack))
		return errors.New("abort the transaction")
	})
	require.Error(t, err)

	var count int
	require.NoError(t, suite.SQLX.GetContext(
		ctx,
		&count,
		`SELECT COUNT(*) FROM admin_audit_log WHERE id = $1`,
		rolledBack.ID,
	))
	require.Zero(t, count)

	committed := rolledBack
	committed.ID = "aud-tx-commit"
	require.NoError(t, tx.WithinTransaction(ctx, func(ctx context.Context) error {
		return recorder.Record(ctx, committed)
	}))

	require.NoError(t, suite.SQLX.GetContext(
		ctx,
		&count,
		`SELECT COUNT(*) FROM admin_audit_log WHERE id = $1`,
		committed.ID,
	))
	require.Equal(t, 1, count)
}

func TestAdminAuditLog_IsAppendOnly(t *testing.T) {
	suite, recorder := setupRecorder(t)
	ctx := context.Background()

	require.NoError(t, recorder.Record(ctx, audit.Entry{
		ID:         "aud-immutable",
		ActorID:    "usr-actor",
		Action:     "user.status.update",
		TargetType: "user",
		TargetID:   "usr-target",
	}))

	_, err := suite.SQLX.ExecContext(
		ctx,
		`UPDATE admin_audit_log SET reason = 'rewritten' WHERE id = $1`,
		"aud-immutable",
	)
	require.ErrorContains(t, err, "append-only")

	_, err = suite.SQLX.ExecContext(
		ctx,
		`DELETE FROM admin_audit_log WHERE id = $1`,
		"aud-immutable",
	)
	require.ErrorContains(t, err, "append-only")
}

func TestPostgresRecorder_ValidatesEntryBeforeWriting(t *testing.T) {
	// Validation happens before the executor is used, so a zero repository is
	// enough here and no database is needed.
	recorder := audit.NewPostgresRecorder(appdb.Repository{}, &generator.IDGenerator{})
	ctx := context.Background()

	require.Error(t, recorder.Record(ctx, audit.Entry{
		Action:     "user.status.update",
		TargetType: "user",
		TargetID:   "usr-target",
	}))
	require.Error(t, recorder.Record(ctx, audit.Entry{
		ActorID:    "usr-actor",
		TargetType: "user",
		TargetID:   "usr-target",
	}))
	require.Error(t, recorder.Record(ctx, audit.Entry{
		ActorID: "usr-actor",
		Action:  "user.status.update",
	}))
	require.Error(t, recorder.Record(ctx, audit.Entry{
		ActorID:    "usr-actor",
		Action:     "user.status.update",
		TargetType: "user",
		TargetID:   "usr-target",
		IPAddress:  "not-an-ip",
	}))
}
