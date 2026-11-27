package app

import (
	"fmt"

	"github.com/AppeiYA/requisition-system/internal/attachment"
	attachmenthttp "github.com/AppeiYA/requisition-system/internal/attachment/adapters/in/http"
	attachmentworker "github.com/AppeiYA/requisition-system/internal/attachment/adapters/in/worker"
	attachmentexternal "github.com/AppeiYA/requisition-system/internal/attachment/adapters/out/external"
	attachmentpostgres "github.com/AppeiYA/requisition-system/internal/attachment/adapters/out/postgres"
	requisitionadapter "github.com/AppeiYA/requisition-system/internal/attachment/adapters/out/requisition"
	"github.com/AppeiYA/requisition-system/internal/attachment/adapters/out/storage/local"
	s3adapter "github.com/AppeiYA/requisition-system/internal/attachment/adapters/out/storage/s3"
	attachmenttelemetry "github.com/AppeiYA/requisition-system/internal/attachment/adapters/out/telemetry"
	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
	attachmentusecase "github.com/AppeiYA/requisition-system/internal/attachment/usecase"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/mime"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/telemetry"
)

// attachmentRepositories are shared by the creation-time registrar and the
// HTTP attachment use cases so both observe the same transaction context.
type attachmentRepositories struct {
	attachments   *attachmentpostgres.AttachmentRepository
	stagedUploads *attachmentpostgres.StagedUploadRepository
	outbox        *attachmentpostgres.ReconciliationOutboxRepository
}

func newAttachmentRepositories(baseRepo db.Repository) *attachmentRepositories {
	return &attachmentRepositories{
		attachments:   attachmentpostgres.NewAttachmentRepository(baseRepo),
		stagedUploads: attachmentpostgres.NewStagedUploadRepository(baseRepo),
		outbox:        attachmentpostgres.NewReconciliationOutboxRepository(baseRepo),
	}
}

// buildAttachmentRegistrar is deliberately separate from buildAttachment:
// requisition creation needs this port before the attachment access checker
// can be built from the requisition port, which would otherwise be a cycle.
func (a *App) buildAttachmentRegistrar(
	repos *attachmentRepositories,
	idGen attachmentout.IDGenerator,
	maxPerRef int,
) requisitionout.StagedAttachmentRegistrar {
	claim := attachmentusecase.NewClaimStagedAttachmentsUsecase(
		repos.stagedUploads, repos.attachments, repos.outbox, idGen, maxPerRef,
	)
	return requisitionadapter.NewStagedAttachmentRegistrar(claim)
}

// newAttachmentStorage is the single place that chooses a storage driver.
// Both the HTTP module and the worker call it so they always agree on where
// staged and permanent objects live.
func newAttachmentStorage(cfg *config.Config) (attachmentout.FileStorage, error) {
	switch cfg.Attachment.StorageDriver {
	case "local":
		return local.New(cfg.Attachment.LocalRoot), nil
	case "s3":
		return s3adapter.New(s3adapter.Options{
			Bucket:          cfg.Attachment.S3Bucket,
			Region:          cfg.Attachment.S3Region,
			Endpoint:        cfg.Attachment.S3Endpoint,
			AccessKeyID:     cfg.Attachment.S3AccessKeyID,
			SecretAccessKey: cfg.Attachment.S3SecretAccessKey,
			ForcePathStyle:  cfg.Attachment.S3ForcePathStyle,
			UseSSL:          cfg.Attachment.S3UseSSL,
		})
	default:
		return nil, fmt.Errorf("unsupported attachment storage driver %q", cfg.Attachment.StorageDriver)
	}
}

// buildAttachmentWorker wires the outbox reconciler with the same storage and
// repositories used by the API. It is a package-level helper because both the
// API and worker composition roots need to share the storage choice.
func buildAttachmentWorker(
	cfg *config.Config,
	repos *attachmentRepositories,
	txManager attachmentout.TransactionManager,
	operationTelemetry *telemetry.OperationTelemetry,
) (*attachment.AttachmentWorkerModule, error) {
	fileStorage, err := newAttachmentStorage(cfg)
	if err != nil {
		return nil, err
	}

	metrics, err := attachmenttelemetry.NewRecorder()
	if err != nil {
		return nil, err
	}

	reconcile := attachmentusecase.NewReconcileOutboxUsecase(
		repos.outbox, repos.attachments, fileStorage, txManager, metrics,
		cfg.Attachment.ReconcileBatchSize,
		cfg.Attachment.ReconcileMaxAttempts,
		cfg.Attachment.ReconcileBackoff,
	)
	handler := attachmentworker.NewReconcileOutboxHandler(reconcile, operationTelemetry)
	cleanup := attachmentusecase.NewCleanupAttachmentsUsecase(
		repos.stagedUploads, repos.attachments, repos.outbox, fileStorage, txManager, metrics,
		cfg.Attachment.ReconcileBatchSize,
		cfg.Attachment.ReconcileStuckAfter,
	)
	cleanupHandler := attachmentworker.NewCleanupAttachmentsHandler(cleanup, operationTelemetry)
	return attachment.NewAttachmentWorkerModule(handler, cleanupHandler), nil
}

// buildAttachment wires the attachment use cases to their adapters. It is the
// only place that chooses a storage driver and discovers the optional signed
// URL capability; domain, use cases and handlers never see either.
func (a *App) buildAttachment(
	cfg *config.Config,
	idGen attachmentout.IDGenerator,
	accessChecker attachmentout.RefAccessChecker,
	txManager attachmentout.TransactionManager,
	repos *attachmentRepositories,
) (*attachment.AttachmentModule, error) {
	fileStorage, err := newAttachmentStorage(cfg)
	if err != nil {
		return nil, err
	}

	var signer attachmentout.SignedURLProvider
	if provider, ok := fileStorage.(attachmentout.SignedURLProvider); ok {
		signer = provider
	}

	repo := repos.attachments
	stagedRepo := repos.stagedUploads
	detector := mime.New()

	uploadUC := attachmentusecase.NewUploadAttachmentUsecase(
		repo, fileStorage, detector, accessChecker, idGen, txManager,
		cfg.Attachment.MaxSizeBytes, cfg.Attachment.MaxPerRef,
	)
	stageUC := attachmentusecase.NewStageUploadUsecase(
		stagedRepo, fileStorage, detector, idGen, txManager,
		cfg.Attachment.MaxSizeBytes, cfg.Attachment.StagingTTL,
		int64(cfg.Attachment.StagingMaxPendingPerUser), cfg.Attachment.StagingMaxBytesPerUser,
	)
	downloadUC := attachmentusecase.NewDownloadAttachmentUsecase(
		repo, fileStorage, accessChecker, signer, cfg.Attachment.SignedURLTTL,
	)
	getUC := attachmentusecase.NewGetAttachmentUsecase(repo, accessChecker)
	listUC := attachmentusecase.NewListAttachmentsUsecase(repo, accessChecker)
	deleteUC := attachmentusecase.NewDeleteAttachmentUsecase(repo, fileStorage, accessChecker, txManager)
	deleteForRefUC := attachmentusecase.NewDeleteAttachmentsForRefUsecase(repo, fileStorage, txManager)

	portIn := attachmentin.NewAttachmentPortIn(
		uploadUC, stageUC, downloadUC, getUC, listUC, deleteUC, deleteForRefUC,
	)
	handler := attachmenthttp.NewAttachmentHandler(
		portIn, cfg.Attachment.MaxSizeBytes, cfg.API.BaseURL,
	)

	return attachment.NewAttachmentModule(portIn, handler), nil
}

// Compile-time guard: the concrete access checker is built over the
// requisition module's port in, so no attachment layer imports requisition
// domain types.
var _ attachmentout.RefAccessChecker = (*attachmentexternal.RequisitionAccessChecker)(nil)
