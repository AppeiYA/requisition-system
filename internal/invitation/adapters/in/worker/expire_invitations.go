package invitationworker

import (
	"context"
	"log"

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
		log.Printf("expire invitations task failed: %v", err)
		return err
	}

	return nil
}
