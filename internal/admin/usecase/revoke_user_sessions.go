package usecase

import (
	"context"

	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
)

type RevokeUserSessionsUsecase struct {
	users     adminout.UserAdminPort
	telemetry adminout.Telemetry
}

func NewRevokeUserSessionsUsecase(users adminout.UserAdminPort, telemetry adminout.Telemetry) *RevokeUserSessionsUsecase {
	return &RevokeUserSessionsUsecase{users: users, telemetry: telemetry}
}

func (uc *RevokeUserSessionsUsecase) Execute(ctx context.Context, cmd adminin.RevokeUserSessionsCommand) error {
	var operationErr error
	_ = uc.telemetry.Measure(ctx, adminout.OperationRevokeUserSessions, func(ctx context.Context) error {
		operationErr = uc.users.RevokeUserSessions(ctx, cmd.UserID, cmd.Reason, cmd.ActorID)
		return operationErr
	})
	return operationErr
}
