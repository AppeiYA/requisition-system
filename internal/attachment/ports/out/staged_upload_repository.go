package attachmentout

import (
	"context"
	"time"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

// ErrStagedUploadClaimConflict means the token was already claimed, expired or
// removed between the guarded read and the update.
var ErrStagedUploadClaimConflict = customerrors.NewConflictError("staged upload is no longer claimable")

// StagedUploadRepository stores parentless uploads. ClaimByIDs atomically
// claims all matching pending, owned, unexpired rows in one statement.
type StagedUploadRepository interface {
	Save(ctx context.Context, upload *domain.StagedUpload) error
	GetByID(ctx context.Context, id string) (*domain.StagedUpload, error)
	// ClaimByIDs atomically flips matching pending, owned, unexpired rows to
	// claimed and sets claimed_at = now, returning the claimed rows. It must
	// run inside the requisition creation transaction so a partial match rolls
	// everything back.
	ClaimByIDs(
		ctx context.Context,
		ids []string,
		actorID string,
		now time.Time,
	) ([]*domain.StagedUpload, error)
	CountPendingByUser(ctx context.Context, actorID string, now time.Time) (int, int64, error)
	ExpirePending(ctx context.Context, now time.Time, limit int) ([]*domain.StagedUpload, error)
	CountPendingAll(ctx context.Context, now time.Time) (int64, error)
}
