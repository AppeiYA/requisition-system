package usecase

import (
	"context"
	"time"

	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
	"github.com/AppeiYA/requisition-system/internal/admin/tasks"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var (
	ErrUserAlreadyExists  = customerrors.NewConflictError("user with this email already exists")
	ErrPendingInvitation  = customerrors.NewConflictError("pending invitation already exists for this email")
	ErrDepartmentNotFound = customerrors.NewNotFoundError("department not found")
)

type InviteMemberUsecase struct {
	taskQueue               adminout.TaskQueue
	adminIdentityPort       adminout.AdminIdentityPort
	adminInvitationPort     adminout.AdminInvitationPort
	adminDepartmentPort     adminout.AdminDepartmentPort
	adminTransactionManager adminout.TransactionManager
}

func NewInviteMemberUsecase(
	taskQueue adminout.TaskQueue,
	adminIdentityPort adminout.AdminIdentityPort,
	adminInvitationPort adminout.AdminInvitationPort,
	adminDepartmentPort adminout.AdminDepartmentPort,
	adminTransactionManager adminout.TransactionManager,
) *InviteMemberUsecase {
	return &InviteMemberUsecase{
		taskQueue:               taskQueue,
		adminIdentityPort:       adminIdentityPort,
		adminInvitationPort:     adminInvitationPort,
		adminDepartmentPort:     adminDepartmentPort,
		adminTransactionManager: adminTransactionManager,
	}
}

func (u *InviteMemberUsecase) Execute(ctx context.Context, cmd adminin.InviteMemberCommand) (*adminin.InviteMemberResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	email, err := shareddomain.NewEmail(cmd.Email)
	if err != nil {
		return nil, err
	}

	// The department must exist before we invite anyone into it. The assignment
	// is immutable once the invitation is created.
	department, err := u.adminDepartmentPort.GetDepartmentByID(ctx, cmd.DepartmentID)
	if err != nil {
		return nil, err
	}
	if department == nil {
		return nil, ErrDepartmentNotFound
	}

	if exists, err := u.adminIdentityPort.CheckUserExists(ctx, email); err != nil {
		return nil, err
	} else if exists {
		return nil, ErrUserAlreadyExists
	}

	if exists, err := u.adminInvitationPort.CheckActiveInvitationExists(ctx, email); err != nil {
		return nil, err
	} else if exists {
		return nil, ErrPendingInvitation
	}

	var saved *adminout.SavedInvitation
	err = u.adminTransactionManager.WithinTransaction(ctx, func(ctx context.Context) error {
		var saveErr error
		saved, saveErr = u.adminInvitationPort.SaveInvitation(ctx, adminout.SaveInvitationParams{
			Email:        email,
			FirstName:    cmd.FirstName,
			LastName:     cmd.LastName,
			Phone:        cmd.Phone,
			StaffID:      cmd.StaffID,
			DepartmentID: cmd.DepartmentID,
			JobTitle:     cmd.JobTitle,
			InviterID:    cmd.InviterID,
		})
		return saveErr
	})
	if err != nil {
		return nil, err
	}

	// Delivery is queued only after the invitation is durably persisted: the
	// worker process sends the email and retries on failure. If enqueueing
	// fails, the error surfaces to the caller; the invitation remains valid
	// and can be re-sent.
	if err := u.taskQueue.EnqueueInvitationEmail(ctx, tasks.InvitationEmailPayload{
		To:             email.Value(),
		InvitedBy:      cmd.InviterID,
		Token:          saved.Token,
		ExpiresAt:      saved.ExpiresAt.Format(time.RFC3339),
		DepartmentName: department.Name,
		JobTitle:       cmd.JobTitle,
	}); err != nil {
		return nil, err
	}

	return &adminin.InviteMemberResult{InvitationID: saved.ID}, nil
}
