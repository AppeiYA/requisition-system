package usecase

import (
	"context"
	"time"

	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
	"github.com/AppeiYA/requisition-system/internal/admin/tasks"
)

type ResendInvitationUsecase struct {
	invitations adminout.AdminInvitationManagementPort
	departments adminout.AdminDepartmentPort
	taskQueue   adminout.TaskQueue
	telemetry   adminout.Telemetry
}

func NewResendInvitationUsecase(
	invitations adminout.AdminInvitationManagementPort,
	departments adminout.AdminDepartmentPort,
	taskQueue adminout.TaskQueue,
	telemetry adminout.Telemetry,
) *ResendInvitationUsecase {
	return &ResendInvitationUsecase{
		invitations: invitations,
		departments: departments,
		taskQueue:   taskQueue,
		telemetry:   telemetry,
	}
}

func (uc *ResendInvitationUsecase) Execute(ctx context.Context, cmd adminin.ResendInvitationCommand) (*adminout.ResendInvitationResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var result *adminout.ResendInvitationResult
	var operationErr error
	_ = uc.telemetry.Measure(ctx, adminout.OperationResendInvitation, func(ctx context.Context) error {
		result, operationErr = uc.resend(ctx, cmd)
		return operationErr
	})
	if operationErr != nil {
		return nil, operationErr
	}
	return result, nil
}

func (uc *ResendInvitationUsecase) resend(ctx context.Context, cmd adminin.ResendInvitationCommand) (*adminout.ResendInvitationResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	invitation, err := uc.invitations.GetInvitation(ctx, cmd.InvitationID)
	if err != nil {
		return nil, err
	}
	departmentName, err := uc.departments.GetDepartmentName(ctx, invitation.DepartmentID)
	if err != nil {
		return nil, err
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	result, err := uc.invitations.ResendInvitation(ctx, adminout.ResendInvitationParams{
		InvitationID: cmd.InvitationID,
		ActorID:      cmd.ActorID,
	})
	if err != nil {
		return nil, err
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := uc.taskQueue.EnqueueInvitationEmail(ctx, tasks.InvitationEmailPayload{
		InvitationID:   result.ID,
		To:             result.Email,
		InvitedBy:      result.InvitedBy,
		Token:          result.Token,
		ExpiresAt:      result.ExpiresAt.Format(time.RFC3339),
		DepartmentName: departmentName,
		JobTitle:       result.JobTitle,
	}); err != nil {
		return nil, err
	}
	return result, nil
}
