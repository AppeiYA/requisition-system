package usecase

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var (
	// ErrAttachmentActorRequired means the request carries no authenticated
	// actor (or no role) and must be rejected before any storage is touched.
	ErrAttachmentActorRequired = customerrors.NewUnauthorizedError("attachment actor is required")
	// ErrAttachmentLimitReached means the parent already has its configured
	// number of attachments.
	ErrAttachmentLimitReached = customerrors.NewConflictError("attachment limit reached for this resource")
	// ErrAttachmentTooLarge covers both a lying Content-Length and a body that
	// grows past the configured cap while it is copied.
	ErrAttachmentTooLarge = customerrors.NewCustomError("attachment too large", http.StatusRequestEntityTooLarge)
	// ErrUnsupportedContentType is the 415 returned for bytes outside the
	// domain allow-list. Detector errors that are already custom errors keep
	// their own status instead.
	ErrUnsupportedContentType = customerrors.NewCustomError("unsupported file type", http.StatusUnsupportedMediaType)
	// ErrAttachmentNotFound is returned when no metadata row matches an ID.
	ErrAttachmentNotFound = customerrors.NewNotFoundError("attachment not found")
	// ErrAttachmentIntegrity means the stored object disagrees with its
	// metadata and must not be served.
	ErrAttachmentIntegrity = customerrors.NewInternalServerError("attachment integrity check failed")
)

var _ attachmentin.UploadAttachmentUsecaseInt = (*UploadAttachmentUsecase)(nil)

// UploadAttachmentUsecase stores file bytes first and metadata second.
type UploadAttachmentUsecase struct {
	repo         attachmentout.AttachmentRepository
	storage      attachmentout.FileStorage
	detector     attachmentout.ContentTypeDetector
	access       attachmentout.RefAccessChecker
	idGen        attachmentout.IDGenerator
	tx           attachmentout.TransactionManager
	maxSizeBytes int64
	maxPerRef    int
}

func NewUploadAttachmentUsecase(
	repo attachmentout.AttachmentRepository,
	storage attachmentout.FileStorage,
	detector attachmentout.ContentTypeDetector,
	access attachmentout.RefAccessChecker,
	idGen attachmentout.IDGenerator,
	tx attachmentout.TransactionManager,
	maxSizeBytes int64,
	maxPerRef int,
) *UploadAttachmentUsecase {
	return &UploadAttachmentUsecase{
		repo:         repo,
		storage:      storage,
		detector:     detector,
		access:       access,
		idGen:        idGen,
		tx:           tx,
		maxSizeBytes: maxSizeBytes,
		maxPerRef:    maxPerRef,
	}
}

func (uc *UploadAttachmentUsecase) Execute(ctx context.Context, cmd attachmentin.UploadAttachmentCommand) (*domain.Attachment, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if cmd.ActorID == "" || cmd.ActorRole == "" {
		return nil, ErrAttachmentActorRequired
	}

	ref, err := domain.NewAttachmentRef(domain.AttachmentRefName(cmd.RefName), cmd.RefID)
	if err != nil {
		return nil, err
	}
	fileName, err := domain.NewAttachmentFileName(cmd.FileName)
	if err != nil {
		return nil, err
	}
	refName, refID := ref.RefName().String(), ref.ID()

	if err := uc.access.CanAccess(ctx, refName, refID, cmd.ActorID, cmd.ActorRole, attachmentout.ActionUpload); err != nil {
		return nil, err
	}

	count, err := uc.repo.CountByRef(ctx, refName, refID)
	if err != nil {
		return nil, err
	}
	if count >= uc.maxPerRef {
		return nil, ErrAttachmentLimitReached
	}

	if cmd.Size > uc.maxSizeBytes {
		return nil, ErrAttachmentTooLarge
	}
	if cmd.Body == nil {
		return nil, ErrUnsupportedContentType
	}

	contentType, replay, err := detectUploadBody(cmd.Body, uc.detector)
	if err != nil {
		return nil, err
	}

	attachmentID, err := uc.idGen.GenerateID(domain.AttachmentIDPrefix)
	if err != nil {
		return nil, err
	}
	key, err := domain.NewAttachmentKey(ref, attachmentID)
	if err != nil {
		return nil, err
	}

	// The declared size is only a hint: cap what is actually copied and rely on
	// Put to compute the SHA-256 checksum from the bytes actually written.
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
		uc.abortUpload(ctx, key, attachmentID, refName, refID, ErrAttachmentTooLarge)
		return nil, ErrAttachmentTooLarge
	}

	attachment, err := domain.NewAttachment(
		attachmentID,
		ref,
		key,
		fileName,
		contentType,
		storedSize,
		checksum,
		cmd.ActorID,
	)
	if err != nil {
		uc.abortUpload(ctx, key, attachmentID, refName, refID, err)
		return nil, err
	}

	if err := uc.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		return uc.repo.Save(txCtx, attachment)
	}); err != nil {
		uc.abortUpload(ctx, key, attachmentID, refName, refID, err)
		return nil, err
	}

	return attachment, nil
}

// abortUpload removes an object whose metadata could not be committed. The
// original failure is returned to the caller by the caller; this only logs.
func (uc *UploadAttachmentUsecase) abortUpload(ctx context.Context, key, attachmentID, refName, refID string, cause error) {
	slog.Warn("attachment metadata write failed",
		"attachment_id", attachmentID,
		"ref_name", refName,
		"ref_id", refID,
		"error", sanitisedAttachmentError(cause),
	)
	if err := uc.storage.Delete(ctx, key); err != nil {
		slog.Warn("failed to remove orphaned attachment object",
			"attachment_id", attachmentID,
			"ref_name", refName,
			"ref_id", refID,
			"error", sanitisedAttachmentError(err),
		)
	}
}

// sanitisedAttachmentError keeps driver details (paths, signed URLs, SQL)
// out of logs while still recording which class of error occurred.
func sanitisedAttachmentError(err error) string {
	if err == nil {
		return ""
	}
	var customErr *customerrors.CustomError
	if errors.As(err, &customErr) {
		return customErr.Message
	}
	return "unexpected error"
}

// attachmentRecordError normalises a missing metadata row to one 404 error and
// leaves every other repository error untouched.
func attachmentRecordError(err error, attachment *domain.Attachment) error {
	if err != nil {
		var customErr *customerrors.CustomError
		if errors.As(err, &customErr) && customErr.Code == http.StatusNotFound {
			return ErrAttachmentNotFound
		}
		return err
	}
	if attachment == nil {
		return ErrAttachmentNotFound
	}
	return nil
}
