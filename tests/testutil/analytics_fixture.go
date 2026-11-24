package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Analytics fixture identities. Tests use these when generating admin/user
// access tokens or when asserting department-scoped results.
const (
	AnalyticsAdminID                 = "usr-analytics-admin"
	AnalyticsUserID                  = "usr-analytics-user"
	AnalyticsSuspendedUserID         = "usr-analytics-suspended"
	AnalyticsInactiveUserID          = "usr-analytics-inactive"
	AnalyticsEngineeringDepartmentID = "dpt-analytics-engineering"
	AnalyticsFinanceDepartmentID     = "dpt-analytics-finance"
)

// AnalyticsFixture describes the fixed data set seeded by
// SeedAnalyticsFixture. The window covers every in-range row and excludes
// the deliberately old user, invitation and requisition.
type AnalyticsFixture struct {
	Anchor  time.Time
	From    time.Time
	To      time.Time
	AdminID string
	UserID  string
}

// SeedAnalyticsFixture inserts a deterministic analytics data set in one
// transaction so the deferred requisition status-sync constraints are
// satisfied at commit time.
func SeedAnalyticsFixture(t *testing.T, suite *Suite) *AnalyticsFixture {
	t.Helper()

	anchor := time.Now().UTC().Truncate(24 * time.Hour)
	at := func(days, hour int) time.Time {
		return anchor.AddDate(0, 0, days).Add(time.Duration(hour) * time.Hour)
	}
	day := func(days int) string {
		return at(days, 0).Format("2006-01-02")
	}

	ctx := context.Background()
	tx, err := suite.SQLX.BeginTxx(ctx, nil)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()

	exec := func(query string, args ...any) {
		t.Helper()
		_, err := tx.ExecContext(ctx, query, args...)
		require.NoError(t, err)
	}

	// Users: 4 total, 2 active, 1 suspended, 1 inactive, 1 admin,
	// 3 created inside the reporting window.
	exec(`
		INSERT INTO users (id, first_name, last_name, email, password_hash, status, role, created_at)
		VALUES
		  ($1, 'Ada', 'Lovelace', 'analytics.admin@example.com', 'analytics-password-hash', 'active',  'admin', $2),
		  ($3, 'Grace', 'Hopper', 'analytics.user@example.com', 'analytics-password-hash', 'active',  'user',  $4),
		  ($5, 'Alan', 'Turing', 'analytics.suspended@example.com', 'analytics-password-hash', 'suspended', 'user', $6),
		  ($7, 'Katherine', 'Johnson', 'analytics.inactive@example.com', 'analytics-password-hash', 'inactive', 'user', $8)
	`,
		AnalyticsAdminID, at(-5, 9),
		AnalyticsUserID, at(-4, 9),
		AnalyticsSuspendedUserID, at(-10, 9),
		AnalyticsInactiveUserID, at(-40, 9),
	)

	exec(`
		INSERT INTO departments (id, name, description, code, status, created_at)
		VALUES
		  ($1, 'Engineering', 'Analytics engineering', 'AENG', 'active', $2),
		  ($3, 'Finance', 'Analytics finance', 'AFIN', 'active', $4)
	`,
		AnalyticsEngineeringDepartmentID, at(-20, 8),
		AnalyticsFinanceDepartmentID, at(-10, 8),
	)

	exec(`
		INSERT INTO department_memberships (user_id, department_id, job_title, status, created_at)
		VALUES
		  ($1, $2, 'Engineering Lead', 'active', $3),
		  ($4, $2, 'Engineer', 'active', $5),
		  ($6, $7, 'Analyst', 'suspended', $8)
	`,
		AnalyticsAdminID, AnalyticsEngineeringDepartmentID, at(-20, 8),
		AnalyticsUserID, at(-19, 8),
		AnalyticsSuspendedUserID, AnalyticsFinanceDepartmentID, at(-10, 8),
	)

	// 5 invitations in range (one of each lifecycle state) and one revoked
	// invitation outside it. Accepted took 48 hours.
	exec(`
		INSERT INTO invitations (
			id, first_name, last_name, email, token_hash, invited_by,
			expires_at, accepted_at, revoked_at, delivered_at, delivery_failed_at,
			department_id, job_title, status, created_at
		)
		VALUES
		  ($1, 'Invited', 'Accepted', 'analytics.accepted@example.com', 'hash-1', $2, $3, $4, NULL, NULL, NULL, $5, 'Engineer', 'accepted', $6),
		  ($7, 'Invited', 'Delivered', 'analytics.delivered@example.com', 'hash-2', $2, $3, NULL, NULL, $8, NULL, $5, 'Engineer', 'delivered', $9),
		  ($10, 'Invited', 'Pending', 'analytics.pending@example.com', 'hash-3', $2, $3, NULL, NULL, NULL, NULL, $11, 'Analyst', 'pending', $12),
		  ($13, 'Invited', 'Expired', 'analytics.expired@example.com', 'hash-4', $2, $14, NULL, NULL, NULL, NULL, $11, 'Analyst', 'expired', $15),
		  ($16, 'Invited', 'Failed', 'analytics.failed@example.com', 'hash-5', $2, $3, NULL, NULL, NULL, $17, $5, 'Engineer', 'failed', $18),
		  ($19, 'Invited', 'Revoked', 'analytics.revoked@example.com', 'hash-6', $2, $20, NULL, $21, NULL, NULL, $5, 'Engineer', 'revoked', $22)
	`,
		"inv-analytics-1", AnalyticsAdminID, at(7, 0), at(-3, 9), AnalyticsEngineeringDepartmentID, at(-5, 9),
		"inv-analytics-2", at(-4, 10), at(-4, 9),
		"inv-analytics-3", AnalyticsFinanceDepartmentID, at(-2, 9),
		"inv-analytics-4", at(-1, 0), at(-6, 9),
		"inv-analytics-5", at(-1, 10), at(-1, 9),
		"inv-analytics-6", at(-33, 0), at(-39, 0), at(-40, 9),
	)

	exec(`
		INSERT INTO item_categories (id, name, description, status, created_at)
		VALUES
		  ($1, 'Office equipment', 'Analytics equipment', 'active', $2),
		  ($3, 'Furniture', 'Analytics furniture', 'inactive', $4)
	`,
		"cat-analytics-equipment", at(-20, 8),
		"cat-analytics-furniture", at(-20, 8),
	)

	exec(`
		INSERT INTO items (id, name, description, category_id, unit_cost, status, created_at, updated_at)
		VALUES
		  ($1, 'Laptop', 'Analytics laptop', 'cat-analytics-equipment', 500000, 'active', $2, $2),
		  ($3, 'Notebook', 'Analytics notebook', 'cat-analytics-equipment', 10000, 'active', $4, $4),
		  ($5, 'Desk', 'Analytics desk', 'cat-analytics-furniture', 250000, 'inactive', $6, $6)
	`,
		"itm-analytics-laptop", at(-20, 8),
		"itm-analytics-notebook", at(-20, 8),
		"itm-analytics-desk", at(-20, 8),
	)

	exec(`
		INSERT INTO requested_items (
			id, name, category_id, description, created_by, review_status,
			promoted_item_id, reviewed_by, reviewed_at, created_at
		)
		VALUES
		  ($1, 'Standing desk', 'cat-analytics-furniture', 'Analytics desk request', $2, 'unreviewed', NULL, NULL, NULL, $3),
		  ($4, 'Ergonomic chair', 'cat-analytics-furniture', 'Analytics chair request', $2, 'unreviewed', NULL, NULL, NULL, $5),
		  ($6, 'Monitor', 'cat-analytics-equipment', 'Analytics monitor request', $2, 'promoted', $7, $8, $9, $10),
		  ($11, 'Whiteboard', 'cat-analytics-equipment', 'Analytics whiteboard request', $2, 'dismissed', NULL, $8, $12, $13)
	`,
		"ritem-analytics-desk", AnalyticsUserID, at(-10, 9),
		"ritem-analytics-chair", at(-35, 9),
		"ritem-analytics-monitor", "itm-analytics-laptop", AnalyticsAdminID, at(-2, 9), at(-3, 9),
		"ritem-analytics-board", at(-1, 9), at(-2, 9),
	)

	// Five in-range requisitions plus one outside the window. History is
	// written in the same transaction to satisfy the status-sync trigger.
	exec(`
		INSERT INTO requisitions (
			id, reference_code, requested_by, department_id, needed_by,
			justification, priority, status, created_at, updated_at
		)
		VALUES
		  ($1,  'REQ-ANALYTICS-1', $2,  $3,  $4,  'Analytics requisition 1', 'high',   'pending',   $5,  $5),
		  ($6,  'REQ-ANALYTICS-2', $7,  $3,  $8,  'Analytics requisition 2', 'medium', 'approved',  $9,  $9),
		  ($10, 'REQ-ANALYTICS-3', $2,  $11, $12, 'Analytics requisition 3', 'high',   'completed', $13, $13),
		  ($14, 'REQ-ANALYTICS-4', $7,  $3,  $15, 'Analytics requisition 4', 'low',    'rejected',  $16, $16),
		  ($17, 'REQ-ANALYTICS-5', $7,  $11, $18, 'Analytics requisition 5', 'medium', 'pending',   $19, $19),
		  ($20, 'REQ-ANALYTICS-6', $7,  $3,  $21, 'Analytics requisition 6', 'low',    'pending',   $22, $22)
	`,
		"req-analytics-1", AnalyticsAdminID, AnalyticsEngineeringDepartmentID, day(5), at(-2, 9),
		"req-analytics-2", AnalyticsUserID, day(-1), at(-1, 10),
		"req-analytics-3", AnalyticsFinanceDepartmentID, day(2), at(-3, 11),
		"req-analytics-4", day(3), at(-2, 14),
		"req-analytics-5", day(-2), at(-40, 9),
		"req-analytics-6", day(-1), at(-1, 15),
	)

	exec(`
		INSERT INTO requisition_status_history (
			id, requisition_id, sequence, from_status, to_status,
			acted_by, actor_role, created_at
		)
		VALUES
		  ('rsh-analytics-1-1', 'req-analytics-1', 1, NULL, 'pending',   $1, 'staff', $2),
		  ('rsh-analytics-2-1', 'req-analytics-2', 1, NULL, 'pending',   $3, 'staff', $4),
		  ('rsh-analytics-2-2', 'req-analytics-2', 2, 'pending', 'approved', $1, 'admin', $5),
		  ('rsh-analytics-3-1', 'req-analytics-3', 1, NULL, 'pending',   $1, 'staff', $6),
		  ('rsh-analytics-3-2', 'req-analytics-3', 2, 'pending', 'approved', $1, 'admin', $7),
		  ('rsh-analytics-3-3', 'req-analytics-3', 3, 'approved', 'completed', $1, 'admin', $8),
		  ('rsh-analytics-4-1', 'req-analytics-4', 1, NULL, 'pending',   $3, 'staff', $9),
		  ('rsh-analytics-4-2', 'req-analytics-4', 2, 'pending', 'rejected', $1, 'admin', $10),
		  ('rsh-analytics-5-1', 'req-analytics-5', 1, NULL, 'pending',   $3, 'staff', $11),
		  ('rsh-analytics-6-1', 'req-analytics-6', 1, NULL, 'pending',   $3, 'staff', $12)
	`,
		AnalyticsAdminID, at(-2, 9),
		AnalyticsUserID, at(-1, 10), at(-1, 11),
		at(-3, 11), at(-3, 12), at(-3, 14),
		at(-2, 14), at(-2, 18),
		at(-40, 9), at(-1, 15),
	)

	exec(`
		INSERT INTO requisition_items (
			id, requisition_id, line_number, item_id, requested_item_id,
			item_name_snapshot, quantity, estimated_unit_amount, created_at
		)
		VALUES
		  ('reqitem-analytics-1-1', 'req-analytics-1', 1, 'itm-analytics-laptop', NULL, 'Laptop', 2, 500000, $1),
		  ('reqitem-analytics-1-2', 'req-analytics-1', 2, NULL, 'ritem-analytics-desk', 'Standing desk', 1, 250000, $1),
		  ('reqitem-analytics-2-1', 'req-analytics-2', 1, 'itm-analytics-notebook', NULL, 'Notebook', 5, 10000, $2),
		  ('reqitem-analytics-3-1', 'req-analytics-3', 1, 'itm-analytics-laptop', NULL, 'Laptop', 1, 500000, $3),
		  ('reqitem-analytics-6-1', 'req-analytics-6', 1, NULL, 'ritem-analytics-chair', 'Ergonomic chair', 2, 150000, $4)
	`,
		at(-2, 9), at(-1, 10), at(-3, 11), at(-1, 15),
	)

	require.NoError(t, tx.Commit())

	return &AnalyticsFixture{
		Anchor:  anchor,
		From:    at(-30, 0),
		To:      at(1, 0),
		AdminID: AnalyticsAdminID,
		UserID:  AnalyticsUserID,
	}
}
