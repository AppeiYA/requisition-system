package attachmentout

import (
	"context"
	"time"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

// ErrAttachmentExists covers both a duplicate storage key and a second claim
// of the same staged upload. Both mean a creation-time attachment conflict.
var ErrAttachmentExists = customerrors.NewConflictError("attachment already exists")

type AttachmentRepository interface {
	Save(ctx context.Context, attachment *domain.Attachment) error
	// UpdateStatus persists a lifecycle transition decided by the aggregate.
	UpdateStatus(ctx context.Context, id string, status domain.AttachmentStatus) error
	GetByID(ctx context.Context, id string) (*domain.Attachment, error)
	ListByRef(ctx context.Context, refName, refID string, limit, offset int) ([]*domain.Attachment, int, error)
	CountByRef(ctx context.Context, refName, refID string) (int, error)
	Delete(ctx context.Context, id string) error
	DeleteByRef(ctx context.Context, refName, refID string) ([]string, error)
	ResetStuckProcessing(ctx context.Context, stuckBefore time.Time, limit int) ([]string, error)
}
