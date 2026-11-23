package postgres_test

import (
	"context"
	"testing"
	"time"

	invitationpostgres "github.com/AppeiYA/requisition-system/internal/invitation/adapters/out/postgres"
	invitationdomain "github.com/AppeiYA/requisition-system/internal/invitation/domain"
	appdb "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/tests/testutil"
	"github.com/stretchr/testify/require"
)

func seedAdminInvitationRows(t *testing.T, suite *testutil.Suite) {
	t.Helper()
	ctx := context.Background()
	require.NoError(t, suite.ResetDatabase(ctx))

	_, err := suite.DB.Pool.Exec(ctx, `
		INSERT INTO users (id, first_name, last_name, email, password_hash, status, role)
		VALUES
			('usr-repo-admin', 'Repo', 'Admin', 'repo-admin@example.com', 'hash', 'active', 'admin'),
			('usr-repo-other', 'Other', 'Inviter', 'repo-other@example.com', 'hash', 'active', 'user');

		INSERT INTO departments (id, name, description, code, status)
		VALUES
			('dpt-repo-a', 'Finance', 'Finance department', 'FIN', 'active'),
			('dpt-repo-b', 'Logistics', 'Logistics department', 'LOG', 'active');

		INSERT INTO invitations (
			id, first_name, last_name, email, department_id, job_title,
			token_hash, invited_by, status, expires_at, accepted_at,
			revoked_at, delivered_at, delivery_failed_at, created_at
		) VALUES
			('inv-repo-one', 'Alice', 'Stone', 'repo-one@example.com', 'dpt-repo-a', 'Engineer',
			 'hash-one', 'usr-repo-admin', 'pending', '2099-01-01T00:00:00Z', NULL,
			 NULL, NULL, NULL, '2026-01-02T00:00:00Z'),
			('inv-repo-two', 'Bob', 'Stone', 'repo-two@example.com', 'dpt-repo-a', 'Analyst',
			 'hash-two', 'usr-repo-admin', 'accepted', '2099-01-01T00:00:00Z', '2026-01-04T00:00:00Z',
			 NULL, NULL, NULL, '2026-01-03T00:00:00Z'),
			('inv-repo-three', 'Carol', 'Stone', 'repo-three@example.com', 'dpt-repo-b', 'Manager',
			 'hash-three', 'usr-repo-other', 'failed', '2099-01-01T00:00:00Z', NULL,
			 NULL, NULL, '2026-01-05T00:00:00Z', '2026-01-04T00:00:00Z'),
			('inv-repo-four', 'Dave', 'Stone', 'repo-four@example.com', 'dpt-repo-b', 'Designer',
			 'hash-four', 'usr-repo-other', 'delivered', '2099-01-01T00:00:00Z', NULL,
			 NULL, '2026-01-06T00:00:00Z', NULL, '2026-01-05T00:00:00Z');
	`)
	require.NoError(t, err)
}

func TestAdminInvitationRepositoryListAndCount(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	seedAdminInvitationRows(t, suite)

	repository := invitationpostgres.NewInvitationRepository(appdb.NewRepository(suite.AppDB))
	ctx := context.Background()

	baseFilter := invitationdomain.InvitationListFilter{
		Limit:  100,
		Offset: 0,
	}

	tests := []struct {
		name      string
		mutate    func(*invitationdomain.InvitationListFilter)
		wantIDs   []string
		wantTotal int
	}{
		{
			name:      "defaults order by newest and paginate",
			mutate:    func(f *invitationdomain.InvitationListFilter) { f.Limit = 2 },
			wantIDs:   []string{"inv-repo-four", "inv-repo-three"},
			wantTotal: 4,
		},
		{
			name:      "filters by case-insensitive search",
			mutate:    func(f *invitationdomain.InvitationListFilter) { f.Search = "BO" },
			wantIDs:   []string{"inv-repo-two"},
			wantTotal: 1,
		},
		{
			name:      "filters by status",
			mutate:    func(f *invitationdomain.InvitationListFilter) { f.Status = "pending" },
			wantIDs:   []string{"inv-repo-one"},
			wantTotal: 1,
		},
		{
			name:      "filters by department",
			mutate:    func(f *invitationdomain.InvitationListFilter) { f.DepartmentID = "dpt-repo-a" },
			wantIDs:   []string{"inv-repo-two", "inv-repo-one"},
			wantTotal: 2,
		},
		{
			name:      "filters by inviter",
			mutate:    func(f *invitationdomain.InvitationListFilter) { f.InvitedBy = "usr-repo-other" },
			wantIDs:   []string{"inv-repo-four", "inv-repo-three"},
			wantTotal: 2,
		},
		{
			name: "filters by inclusive from and exclusive to",
			mutate: func(f *invitationdomain.InvitationListFilter) {
				from := time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)
				to := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
				f.From, f.To = &from, &to
			},
			wantIDs:   []string{"inv-repo-three", "inv-repo-two"},
			wantTotal: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := baseFilter
			tt.mutate(&filter)

			rows, err := repository.ListInvitations(ctx, filter)
			require.NoError(t, err)
			require.Len(t, rows, len(tt.wantIDs))
			for i, wantID := range tt.wantIDs {
				require.Equal(t, wantID, rows[i].ID)
			}

			total, err := repository.CountInvitations(ctx, filter)
			require.NoError(t, err)
			require.Equal(t, tt.wantTotal, total)
		})
	}
}

func TestAdminInvitationRepositoryGetSummaryAndRotate(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	seedAdminInvitationRows(t, suite)

	repository := invitationpostgres.NewInvitationRepository(appdb.NewRepository(suite.AppDB))
	ctx := context.Background()

	summary, err := repository.GetInvitationSummary(ctx, "inv-repo-one")
	require.NoError(t, err)
	require.Equal(t, "inv-repo-one", summary.ID)
	require.Equal(t, "repo-one@example.com", summary.Email)
	require.Equal(t, "Finance", summary.DepartmentName)
	require.Equal(t, "Repo Admin", summary.InviterName)
	require.Equal(t, "pending", summary.Status)

	_, err = repository.GetInvitationSummary(ctx, "inv-missing")
	require.ErrorIs(t, err, invitationpostgres.ErrInvitationNotFoundRepo)

	expiresAt := time.Date(2099, 2, 1, 0, 0, 0, 0, time.UTC)
	require.NoError(t, repository.RotateInvitationToken(ctx, "inv-repo-one", "hash-rotated", expiresAt))

	rotated, err := repository.GetInvitationByTokenHash(ctx, "hash-rotated")
	require.NoError(t, err)
	require.Equal(t, "inv-repo-one", rotated.ID())
	require.Equal(t, expiresAt.UTC(), rotated.ExpiresAt().UTC())

	_, err = repository.GetInvitationByTokenHash(ctx, "hash-one")
	require.Error(t, err)

	require.ErrorIs(t, repository.RotateInvitationToken(ctx, "inv-repo-two", "hash-not-rotated", expiresAt), invitationpostgres.ErrInvitationRotationConflict)
}
