package usecase

import (
	"context"
	"strings"

	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/audit"
)

type RevokeUserSessionsUsecase struct {
	users    identityout.UsersRepository
	sessions identityout.RefreshSessionRepository
	audit    identityout.AuditRecorder
}

func NewRevokeUserSessionsUsecase(
	users identityout.UsersRepository,
	sessions identityout.RefreshSessionRepository,
	auditRecorder identityout.AuditRecorder,
) *RevokeUserSessionsUsecase {
	return &RevokeUserSessionsUsecase{
		users:    users,
		sessions: sessions,
		audit:    auditRecorder,
	}
}

func (uc *RevokeUserSessionsUsecase) Execute(ctx context.Context, cmd identityin.RevokeUserSessionsCommand) error {
	if cmd.UserID == "" {
		return ErrUserIDRequired
	}

	user, err := uc.users.GetUserByID(ctx, cmd.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Revocation is a Redis mutation, so it cannot join a database
	// transaction. The audit entry is written immediately afterwards; losing
	// it would require losing Redis and the audit database at the same time.
	if err := uc.sessions.RevokeAll(ctx, user.ID()); err != nil {
		return err
	}

	return uc.audit.Record(ctx, audit.Entry{
		ActorID:    cmd.ActorID,
		Action:     "user.sessions.revoke",
		TargetType: "user",
		TargetID:   user.ID(),
		Before:     map[string]any{"sessions": "active"},
		After:      map[string]any{"sessions": "revoked"},
		Reason:     strings.TrimSpace(cmd.Reason),
	})
}
