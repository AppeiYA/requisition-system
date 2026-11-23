package identityout

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/audit"
)

// AuditRecorder lets the identity use cases write to the shared append-only
// admin audit log without depending on the audit infrastructure package.
type AuditRecorder interface {
	Record(ctx context.Context, entry audit.Entry) error
}
