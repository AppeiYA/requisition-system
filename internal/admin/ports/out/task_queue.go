package adminout

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/admin/tasks"
)

type TaskQueue interface {
	EnqueueInvitationEmail(ctx context.Context, payload tasks.InvitationEmailPayload) error
}
