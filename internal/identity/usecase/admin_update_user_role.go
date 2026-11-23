package usecase

import (
	"context"
	"strings"

	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/audit"
)

type UpdateUserRoleUsecase struct {
	users     identityout.UsersRepository
	sessions  identityout.RefreshSessionRepository
	txManager identityout.TransactionManager
	audit     identityout.AuditRecorder
}

func NewUpdateUserRoleUsecase(
	users identityout.UsersRepository,
	sessions identityout.RefreshSessionRepository,
	txManager identityout.TransactionManager,
	auditRecorder identityout.AuditRecorder,
) *UpdateUserRoleUsecase {
	return &UpdateUserRoleUsecase{
		users:     users,
		sessions:  sessions,
		txManager: txManager,
		audit:     auditRecorder,
	}
}

func (uc *UpdateUserRoleUsecase) Execute(ctx context.Context, cmd identityin.UpdateUserRoleCommand) error {
	if cmd.UserID == "" {
		return ErrUserIDRequired
	}
	if cmd.ActorID == cmd.UserID {
		return ErrSelfUserChange
	}
	if !shareddomain.IsValidRole(cmd.Role) {
		return ErrInvalidRoleValue
	}
	newRole := shareddomain.Role(cmd.Role)
	reason := strings.TrimSpace(cmd.Reason)

	if err := uc.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Lock the admin set in id order before the target row, so two
		// admins demoting each other cannot deadlock.
		others, err := uc.users.CountOtherActiveAdminsForUpdate(txCtx, cmd.UserID)
		if err != nil {
			return err
		}

		user, err := uc.users.GetUserByIDForUpdate(txCtx, cmd.UserID)
		if err != nil {
			return err
		}
		if user == nil {
			return ErrUserNotFound
		}

		if newRole == shareddomain.RoleAdmin && !user.IsActive() {
			return ErrRoleChangeRequiresActive
		}
		// A demotion of the last active admin would lock everybody out.
		if user.Role() == shareddomain.RoleAdmin && user.IsActive() && newRole == shareddomain.RoleUser && others == 0 {
			return ErrLastActiveAdmin
		}

		before := user.Role().String()
		if err := user.UpdateRole(newRole); err != nil {
			return err
		}
		if err := uc.users.UpdateUserRole(txCtx, user); err != nil {
			return err
		}
		return uc.audit.Record(txCtx, audit.Entry{
			ActorID:    cmd.ActorID,
			Action:     "user.role.update",
			TargetType: "user",
			TargetID:   user.ID(),
			Before:     map[string]any{"role": before},
			After:      map[string]any{"role": newRole.String()},
			Reason:     reason,
		})
	}); err != nil {
		return err
	}

	// The access token's role claim is no longer authoritative, but revoking
	// also forces a fresh token so the claim matches the database immediately.
	return uc.sessions.RevokeAll(ctx, cmd.UserID)
}
