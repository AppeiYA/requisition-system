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
	ErrDepartmentNotFound = customerrors.NewNotFoundError("department not found")
)

type InviteMemberUsecase struct {
	taskQueue               adminout.TaskQueue
	adminIdentityPort       adminout.AdminIdentityPort
	adminInvitationPort     adminout.AdminInvitationPort
	adminDepartmentPort     adminout.AdminDepartmentPort
	adminTransactionManager adminout.TransactionManager
	telemetry               adminout.Telemetry
}

func NewInviteMemberUsecase(
	taskQueue adminout.TaskQueue,
	adminIdentityPort adminout.AdminIdentityPort,
	adminInvitationPort adminout.AdminInvitationPort,
	adminDepartmentPort adminout.AdminDepartmentPort,
	adminTransactionManager adminout.TransactionManager,
	telemetry adminout.Telemetry,
) *InviteMemberUsecase {
	return &InviteMemberUsecase{
		taskQueue:               taskQueue,
		adminIdentityPort:       adminIdentityPort,
		adminInvitationPort:     adminInvitationPort,
		adminDepartmentPort:     adminDepartmentPort,
		adminTransactionManager: adminTransactionManager,
		telemetry:               telemetry,
	}
}

func (u *InviteMemberUsecase) Execute(ctx context.Context, cmd adminin.InviteMemberCommand) (*adminin.InviteMemberResult, error) {
	// A cancelled request never starts the operation, so it is not measured.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var result *adminin.InviteMemberResult
	var operationErr error

	// The measurement covers the whole operation: the invitation is only
	// complete once it is persisted and its delivery is queued. The outcome of
	// the operation is the one reported to the caller, the instrumentation can
	// never replace nor mask it.
	_ = u.telemetry.Measure(ctx, adminout.OperationInviteMember, func(ctx context.Context) error {
		result, operationErr = u.inviteMember(ctx, cmd)
		return operationErr
	})

	if operationErr != nil {
		return nil, operationErr
	}

	return result, nil
}

func (u *InviteMemberUsecase) inviteMember(ctx context.Context, cmd adminin.InviteMemberCommand) (*adminin.InviteMemberResult, error) {
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
		InvitationID:   saved.ID,
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
