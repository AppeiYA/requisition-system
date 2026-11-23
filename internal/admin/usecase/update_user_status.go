package usecase

import (
	"context"

	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
)

type UpdateUserStatusUsecase struct {
	users     adminout.UserAdminPort
	telemetry adminout.Telemetry
}

func NewUpdateUserStatusUsecase(users adminout.UserAdminPort, telemetry adminout.Telemetry) *UpdateUserStatusUsecase {
	return &UpdateUserStatusUsecase{users: users, telemetry: telemetry}
}

func (uc *UpdateUserStatusUsecase) Execute(ctx context.Context, cmd adminin.UpdateUserStatusCommand) error {
	var operationErr error
	_ = uc.telemetry.Measure(ctx, adminout.OperationUpdateUserStatus, func(ctx context.Context) error {
		operationErr = uc.users.UpdateUserStatus(ctx, cmd.UserID, cmd.Status, cmd.Reason, cmd.ActorID)
		return operationErr
	})
	return operationErr
}
