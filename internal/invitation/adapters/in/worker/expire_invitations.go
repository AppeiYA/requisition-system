package invitationworker

import (
	"context"
	"log/slog"

	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
	"github.com/hibiken/asynq"
)

const ExpireInvitations = "invitation:expire"

type ExpireInvitationsHandler struct {
	expireInvitations invitationin.ExpireInvitationsUsecaseInt
}

func NewExpireInvitationsHandler(expireInvitations invitationin.ExpireInvitationsUsecaseInt) *ExpireInvitationsHandler {
	return &ExpireInvitationsHandler{expireInvitations: expireInvitations}
}

func (h *ExpireInvitationsHandler) Handle(ctx context.Context, task *asynq.Task) error {
	if err := h.expireInvitations.Execute(ctx); err != nil {
		slog.ErrorContext(ctx, "expire invitations task failed", slog.String("error.message", err.Error()))
		return err
	}

	return nil
}
