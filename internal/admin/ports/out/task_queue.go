package adminout

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/admin/tasks"
)

// TaskQueue enqueues background work on behalf of the application. It lets
// the invite flow return as soon as the invitation is persisted while the
// slow part (delivering the email) happens in the worker process.
type TaskQueue interface {
	EnqueueInvitationEmail(ctx context.Context, payload tasks.InvitationEmailPayload) error
}
