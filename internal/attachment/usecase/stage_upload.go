package usecase

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

// ErrStagedUploadQuotaExceeded is returned when a user has too many pending
// staged uploads or pending staged bytes.
var ErrStagedUploadQuotaExceeded = customerrors.NewCustomError("staged upload quota exceeded", http.StatusTooManyRequests)

var _ attachmentin.StageUploadUsecaseInt = (*StageUploadUsecase)(nil)

// StageUploadUsecase stores bytes under a parentless staging key and records
// their owner. It never asks RefAccessChecker: the parent does not exist yet,
// and final authorisation happens during requisition creation.
type StageUploadUsecase struct {
	repo              attachmentout.StagedUploadRepository
	storage           attachmentout.FileStorage
	detector          attachmentout.ContentTypeDetector
	idGen             attachmentout.IDGenerator
	tx                attachmentout.TransactionManager
	maxSizeBytes      int64
	ttl               time.Duration
	maxPendingPerUser int
	maxBytesPerUser   int64
	now               func() time.Time
}

func NewStageUploadUsecase(
	repo attachmentout.StagedUploadRepository,
	storage attachmentout.FileStorage,
	detector attachmentout.ContentTypeDetector,
	idGen attachmentout.IDGenerator,
	tx attachmentout.TransactionManager,
	maxSizeBytes int64,
	ttl time.Duration,
	quota ...int64,
) *StageUploadUsecase {
	var maxPending int
	var maxBytes int64
	if len(quota) > 0 {
		maxPending = int(quota[0])
	}
	if len(quota) > 1 {
		maxBytes = quota[1]
	}
	return &StageUploadUsecase{
		repo:              repo,
		storage:           storage,
		detector:          detector,
		idGen:             idGen,
		tx:                tx,
		maxSizeBytes:      maxSizeBytes,
		ttl:               ttl,
		maxPendingPerUser: maxPending,
		maxBytesPerUser:   maxBytes,
		now:               func() time.Time { return time.Now().UTC() },
	}
}

func (uc *StageUploadUsecase) Execute(ctx context.Context, cmd attachmentin.StageUploadCommand) (*domain.StagedUpload, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if cmd.ActorID == "" {
		return nil, ErrAttachmentActorRequired
	}
	fileName, err := domain.NewAttachmentFileName(cmd.FileName)
	if err != nil {
		return nil, err
	}
	if cmd.Size > uc.maxSizeBytes {
		return nil, ErrAttachmentTooLarge
	}

	now := uc.now()
	var pendingBytes int64
	if uc.maxPendingPerUser > 0 || uc.maxBytesPerUser > 0 {
		pendingCount, totalBytes, err := uc.repo.CountPendingByUser(ctx, cmd.ActorID, now)
		if err != nil {
			return nil, err
		}
		if uc.maxPendingPerUser > 0 && pendingCount >= uc.maxPendingPerUser {
			return nil, ErrStagedUploadQuotaExceeded
		}
		if uc.maxBytesPerUser > 0 && totalBytes+cmd.Size > uc.maxBytesPerUser {
			return nil, ErrStagedUploadQuotaExceeded
		}
		pendingBytes = totalBytes
	}

	contentType, replay, err := detectUploadBody(cmd.Body, uc.detector)
	if err != nil {
		return nil, err
	}

	uploadID, err := uc.idGen.GenerateID(domain.StagedUploadIDPrefix)
	if err != nil {
		return nil, err
	}
	key, err := domain.NewStagingUploadKey(uploadID)
	if err != nil {
		return nil, err
	}

	limited := &attachmentSizeLimiter{reader: replay, max: uc.maxSizeBytes}
	info, err := uc.storage.Put(ctx, key, limited, attachmentout.PutOptions{
		ContentType: contentType.String(),
		Size:        cmd.Size,
	})
	if err != nil {
		return nil, err
	}
	checksum := info.Checksum
	storedSize := info.Size
	if storedSize <= 0 {
		storedSize = limited.read
	}
	if storedSize > uc.maxSizeBytes {
		uc.abortStage(ctx, key, uploadID, ErrAttachmentTooLarge)
		return nil, ErrAttachmentTooLarge
	}
	if uc.maxBytesPerUser > 0 && pendingBytes+storedSize > uc.maxBytesPerUser {
		uc.abortStage(ctx, key, uploadID, ErrStagedUploadQuotaExceeded)
		return nil, ErrStagedUploadQuotaExceeded
	}

	now = uc.now()
	upload, err := domain.NewStagedUpload(
		uploadID,
		cmd.ActorID,
		key,
		fileName,
		contentType,
		storedSize,
		checksum,
		uc.ttl,
		now,
	)
	if err != nil {
		uc.abortStage(ctx, key, uploadID, err)
		return nil, err
	}

	if err := uc.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		return uc.repo.Save(txCtx, upload)
	}); err != nil {
		uc.abortStage(ctx, key, uploadID, err)
		return nil, err
	}

	return upload, nil
}

// abortStage removes bytes whose metadata could not be committed and logs a
// sanitised reason. The original error is returned by the caller.
func (uc *StageUploadUsecase) abortStage(ctx context.Context, key, uploadID string, cause error) {
	slog.Warn("staged upload metadata write failed",
		"staged_upload_id", uploadID,
		"error", sanitisedAttachmentError(cause),
	)
	if err := uc.storage.Delete(ctx, key); err != nil {
		slog.Warn("failed to remove orphaned staged upload",
			"staged_upload_id", uploadID,
			"error", sanitisedAttachmentError(err),
		)
	}
}
