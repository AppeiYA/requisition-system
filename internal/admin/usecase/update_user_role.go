package usecase

import (
	"context"

	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
)

type UpdateUserRoleUsecase struct {
	users     adminout.UserAdminPort
	telemetry adminout.Telemetry
}

func NewUpdateUserRoleUsecase(users adminout.UserAdminPort, telemetry adminout.Telemetry) *UpdateUserRoleUsecase {
	return &UpdateUserRoleUsecase{users: users, telemetry: telemetry}
}

func (uc *UpdateUserRoleUsecase) Execute(ctx context.Context, cmd adminin.UpdateUserRoleCommand) error {
	var operationErr error
	_ = uc.telemetry.Measure(ctx, adminout.OperationUpdateUserRole, func(ctx context.Context) error {
		operationErr = uc.users.UpdateUserRole(ctx, cmd.UserID, cmd.Role, cmd.Reason, cmd.ActorID)
		return operationErr
	})
	return operationErr
}
