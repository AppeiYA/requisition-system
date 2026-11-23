package usecase

import (
	"context"
	"strings"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/audit"
)

type UpdateUserStatusUsecase struct {
	users     identityout.UsersRepository
	sessions  identityout.RefreshSessionRepository
	txManager identityout.TransactionManager
	audit     identityout.AuditRecorder
}

func NewUpdateUserStatusUsecase(
	users identityout.UsersRepository,
	sessions identityout.RefreshSessionRepository,
	txManager identityout.TransactionManager,
	auditRecorder identityout.AuditRecorder,
) *UpdateUserStatusUsecase {
	return &UpdateUserStatusUsecase{
		users:     users,
		sessions:  sessions,
		txManager: txManager,
		audit:     auditRecorder,
	}
}

func (uc *UpdateUserStatusUsecase) Execute(ctx context.Context, cmd identityin.UpdateUserStatusCommand) error {
	if cmd.UserID == "" {
		return ErrUserIDRequired
	}
	if cmd.ActorID == cmd.UserID {
		return ErrSelfUserChange
	}

	status, err := domain.NewStatus(cmd.Status)
	if err != nil {
		return err
	}
	reason := strings.TrimSpace(cmd.Reason)
	if status != domain.ACTIVE && reason == "" {
		return ErrReasonRequired
	}

	var before, after string
	if err := uc.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Lock the admin set before the target row. The repository orders
		// those locks by id, so two admins changing each other cannot acquire
		// locks in opposite order and deadlock.
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
		if !user.Status().CanTransitionTo(status) {
			return domain.ErrInvalidStatus
		}

		// Removing the last active admin would lock everybody out.
		if user.IsActive() && user.Role() == shareddomain.RoleAdmin && status != domain.ACTIVE && others == 0 {
			return ErrLastActiveAdmin
		}

		before = user.Status().Value()
		if err := user.UpdateStatus(status); err != nil {
			return err
		}
		after = user.Status().Value()

		if err := uc.users.UpdateUserStatus(txCtx, user); err != nil {
			return err
		}
		return uc.audit.Record(txCtx, audit.Entry{
			ActorID:    cmd.ActorID,
			Action:     "user.status.update",
			TargetType: "user",
			TargetID:   user.ID(),
			Before:     map[string]any{"status": before},
			After:      map[string]any{"status": after},
			Reason:     reason,
		})
	}); err != nil {
		return err
	}

	// Redis cannot join the database transaction. A failure here leaves the
	// status change committed but still blocks the user at the middleware on
	// every request; only refresh tokens remain valid until they expire.
	if status != domain.ACTIVE {
		return uc.sessions.RevokeAll(ctx, cmd.UserID)
	}
	return nil
}
