package attachmentout

import "context"

type Action string

const (
	ActionUpload Action = "upload"
	ActionRead   Action = "read"
	ActionDelete Action = "delete"
)

// RefAccessChecker delegates attachment authorisation to the module that owns
// the parent resource, so attachment rules cannot drift from parent rules.
type RefAccessChecker interface {
	CanAccess(ctx context.Context, refName, refID, actorID, actorRole string, action Action) error
}
