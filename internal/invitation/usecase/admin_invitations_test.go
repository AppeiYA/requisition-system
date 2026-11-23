package usecase_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	invitationdomain "github.com/AppeiYA/requisition-system/internal/invitation/domain"
	invitationmocks "github.com/AppeiYA/requisition-system/internal/invitation/mocks"
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
	invitationusecase "github.com/AppeiYA/requisition-system/internal/invitation/usecase"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
	"github.com/stretchr/testify/require"
)

func adminFixtureInvitation(t *testing.T, status invitationdomain.InvitationStatus) *invitationdomain.Invitation {
	t.Helper()
	invitation := invitationmocks.ValidInvitation()
	switch status {
	case invitationdomain.InvitationStatusAccepted:
		require.NoError(t, invitation.MarkAccepted())
	case invitationdomain.InvitationStatusRevoked:
		require.NoError(t, invitation.MarkAsRevoked())
	case invitationdomain.InvitationStatusDelivered:
		require.NoError(t, invitation.MarkAsDelivered())
	case invitationdomain.InvitationStatusFailed:
		require.NoError(t, invitation.MarkAsFailed())
	case invitationdomain.InvitationStatusPending:
	default:
		t.Fatalf("unsupported fixture status %q", status)
	}
	return invitation
}

func adminFixtureSummary(invitation *invitationdomain.Invitation) *invitationdomain.InvitationSummary {
	now := time.Now()
	return &invitationdomain.InvitationSummary{
		ID:          invitation.ID(),
		Status:      invitation.Status().String(),
		ExpiresAt:   invitation.ExpiresAt(),
		CreatedAt:   invitation.CreatedAt(),
		DeliveredAt: &now,
	}
}

func requireCustomErrorCode(t *testing.T, err error, code int) {
	t.Helper()
	require.Error(t, err)
	var customErr *customerrors.CustomError
	require.ErrorAs(t, err, &customErr)
	require.Equal(t, code, customErr.Code)
}

func TestAdminResendInvitation(t *testing.T) {
	t.Run("rotates the token and audits the change", func(t *testing.T) {
		invitation := adminFixtureInvitation(t, invitationdomain.InvitationStatusDelivered)
		repo := invitationmocks.NewMockInvitationRepository()
		repo.LockInvitationByIDFunc = func(_ context.Context, id string) (*invitationdomain.Invitation, error) {
			require.Equal(t, invitation.ID(), id)
			return invitation, nil
		}
		repo.GetActiveInvitationByEmailFunc = func(_ context.Context, email string) (*invitationdomain.Invitation, error) {
			return invitation, nil
		}
		repo.GetInvitationSummaryFunc = func(_ context.Context, id string) (*invitationdomain.InvitationSummary, error) {
			return adminFixtureSummary(invitation), nil
		}

		var rotatedHash string
		var rotatedExpiry time.Time
		repo.RotateInvitationTokenFunc = func(_ context.Context, id, tokenHash string, expiresAt time.Time) error {
			require.Equal(t, invitation.ID(), id)
			rotatedHash = tokenHash
			rotatedExpiry = expiresAt
			return nil
		}

		auditRecorder := invitationmocks.NewMockAuditRecorder()
		uc := invitationusecase.NewAdminResendInvitationUsecase(repo, invitationmocks.NewMockTransactionManager(), auditRecorder)

		result, err := uc.Execute(context.Background(), invitationin.ResendInvitationCommand{
			InvitationID: invitation.ID(),
			ActorID:      invitationmocks.DefaultInviterID,
		})
		require.NoError(t, err)
		require.Equal(t, invitation.ID(), result.ID)
		require.NotEmpty(t, result.Token)
		require.Equal(t, invitationdomain.HashInvitationToken(result.Token), rotatedHash)
		require.True(t, rotatedExpiry.After(time.Now().Add(23*time.Hour)))
		require.Equal(t, result.ExpiresAt, rotatedExpiry)

		require.Len(t, auditRecorder.Entries, 1)
		entry := auditRecorder.Entries[0]
		require.Equal(t, "invitation.resend", entry.Action)
		require.Equal(t, "invitation", entry.TargetType)
		require.Equal(t, invitation.ID(), entry.TargetID)
		require.Equal(t, invitationmocks.DefaultInviterID, entry.ActorID)
	})

	t.Run("rejects accepted invitations", func(t *testing.T) {
		invitation := adminFixtureInvitation(t, invitationdomain.InvitationStatusAccepted)
		repo := invitationmocks.NewMockInvitationRepository()
		repo.LockInvitationByIDFunc = func(context.Context, string) (*invitationdomain.Invitation, error) {
			return invitation, nil
		}
		auditRecorder := invitationmocks.NewMockAuditRecorder()
		uc := invitationusecase.NewAdminResendInvitationUsecase(repo, invitationmocks.NewMockTransactionManager(), auditRecorder)

		_, err := uc.Execute(context.Background(), invitationin.ResendInvitationCommand{
			InvitationID: invitation.ID(),
			ActorID:      invitationmocks.DefaultInviterID,
		})
		requireCustomErrorCode(t, err, http.StatusConflict)
		require.Empty(t, auditRecorder.Entries)
	})

	t.Run("rejects revoked invitations", func(t *testing.T) {
		invitation := adminFixtureInvitation(t, invitationdomain.InvitationStatusRevoked)
		repo := invitationmocks.NewMockInvitationRepository()
		repo.LockInvitationByIDFunc = func(context.Context, string) (*invitationdomain.Invitation, error) {
			return invitation, nil
		}
		uc := invitationusecase.NewAdminResendInvitationUsecase(repo, invitationmocks.NewMockTransactionManager(), invitationmocks.NewMockAuditRecorder())

		_, err := uc.Execute(context.Background(), invitationin.ResendInvitationCommand{
			InvitationID: invitation.ID(),
			ActorID:      invitationmocks.DefaultInviterID,
		})
		requireCustomErrorCode(t, err, http.StatusConflict)
	})

	t.Run("rejects a second active invitation for the email", func(t *testing.T) {
		invitation := adminFixtureInvitation(t, invitationdomain.InvitationStatusFailed)
		email, err := shareddomain.NewEmail(invitationmocks.DefaultInviteeEmail)
		require.NoError(t, err)
		firstName, err := shareddomain.NewUsername(invitationmocks.DefaultFirstName, shareddomain.FirstName)
		require.NoError(t, err)
		lastName, err := shareddomain.NewUsername(invitationmocks.DefaultLastName, shareddomain.LastName)
		require.NoError(t, err)
		other, err := invitationdomain.NewInvitation(
			"inv-other", firstName, lastName, nil, "", invitationmocks.DefaultDepartmentID,
			invitationmocks.DefaultJobTitle, email, "hash", invitationmocks.DefaultInviterID,
		)
		require.NoError(t, err)
		repo := invitationmocks.NewMockInvitationRepository()
		repo.LockInvitationByIDFunc = func(context.Context, string) (*invitationdomain.Invitation, error) {
			return invitation, nil
		}
		repo.GetActiveInvitationByEmailFunc = func(context.Context, string) (*invitationdomain.Invitation, error) {
			return other, nil
		}
		uc := invitationusecase.NewAdminResendInvitationUsecase(repo, invitationmocks.NewMockTransactionManager(), invitationmocks.NewMockAuditRecorder())

		_, err = uc.Execute(context.Background(), invitationin.ResendInvitationCommand{
			InvitationID: invitation.ID(),
			ActorID:      invitationmocks.DefaultInviterID,
		})
		requireCustomErrorCode(t, err, http.StatusConflict)
	})
}

func TestAdminRevokeInvitation(t *testing.T) {
	t.Run("rejects an oversized reason", func(t *testing.T) {
		uc := invitationusecase.NewAdminRevokeInvitationUsecase(
			invitationmocks.NewMockInvitationRepository(),
			invitationmocks.NewMockTransactionManager(),
			invitationmocks.NewMockAuditRecorder(),
		)

		err := uc.Execute(context.Background(), invitationin.RevokeInvitationCommand{
			InvitationID: "inv-1",
			ActorID:      invitationmocks.DefaultInviterID,
			Reason:       string(make([]byte, 501)),
		})
		requireCustomErrorCode(t, err, http.StatusBadRequest)
	})

	t.Run("revokes and audits a pending invitation", func(t *testing.T) {
		invitation := adminFixtureInvitation(t, invitationdomain.InvitationStatusPending)
		revokedAt := time.Now()
		repo := invitationmocks.NewMockInvitationRepository()
		repo.LockInvitationByIDFunc = func(context.Context, string) (*invitationdomain.Invitation, error) {
			return invitation, nil
		}
		repo.GetInvitationSummaryFunc = func(context.Context, string) (*invitationdomain.InvitationSummary, error) {
			return &invitationdomain.InvitationSummary{
				ID:        invitation.ID(),
				Status:    invitationdomain.InvitationStatusRevoked.String(),
				RevokedAt: &revokedAt,
			}, nil
		}
		revoked := false
		repo.RevokeInvitationFunc = func(_ context.Context, id string) error {
			require.Equal(t, invitation.ID(), id)
			revoked = true
			return nil
		}
		auditRecorder := invitationmocks.NewMockAuditRecorder()
		uc := invitationusecase.NewAdminRevokeInvitationUsecase(repo, invitationmocks.NewMockTransactionManager(), auditRecorder)

		err := uc.Execute(context.Background(), invitationin.RevokeInvitationCommand{
			InvitationID: invitation.ID(),
			ActorID:      invitationmocks.DefaultInviterID,
			Reason:       "duplicate",
		})
		require.NoError(t, err)
		require.True(t, revoked)
		require.Len(t, auditRecorder.Entries, 1)
		require.Equal(t, "invitation.revoke", auditRecorder.Entries[0].Action)
		require.Equal(t, "duplicate", auditRecorder.Entries[0].Reason)
		require.Equal(t, "pending", auditRecorder.Entries[0].Before.(map[string]any)["status"])
	})

	t.Run("rejects an accepted invitation", func(t *testing.T) {
		invitation := adminFixtureInvitation(t, invitationdomain.InvitationStatusAccepted)
		repo := invitationmocks.NewMockInvitationRepository()
		repo.LockInvitationByIDFunc = func(context.Context, string) (*invitationdomain.Invitation, error) {
			return invitation, nil
		}
		uc := invitationusecase.NewAdminRevokeInvitationUsecase(repo, invitationmocks.NewMockTransactionManager(), invitationmocks.NewMockAuditRecorder())

		err := uc.Execute(context.Background(), invitationin.RevokeInvitationCommand{
			InvitationID: invitation.ID(),
			ActorID:      invitationmocks.DefaultInviterID,
		})
		requireCustomErrorCode(t, err, http.StatusConflict)
	})

	t.Run("treats an already revoked invitation as success", func(t *testing.T) {
		invitation := adminFixtureInvitation(t, invitationdomain.InvitationStatusRevoked)
		repo := invitationmocks.NewMockInvitationRepository()
		repo.LockInvitationByIDFunc = func(context.Context, string) (*invitationdomain.Invitation, error) {
			return invitation, nil
		}
		auditRecorder := invitationmocks.NewMockAuditRecorder()
		uc := invitationusecase.NewAdminRevokeInvitationUsecase(repo, invitationmocks.NewMockTransactionManager(), auditRecorder)

		err := uc.Execute(context.Background(), invitationin.RevokeInvitationCommand{
			InvitationID: invitation.ID(),
			ActorID:      invitationmocks.DefaultInviterID,
		})
		require.NoError(t, err)
		require.Empty(t, auditRecorder.Entries)
	})
}
