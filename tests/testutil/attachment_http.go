package testutil

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	attachmenthttp "github.com/AppeiYA/requisition-system/internal/attachment/adapters/in/http"
	attachmentpostgres "github.com/AppeiYA/requisition-system/internal/attachment/adapters/out/postgres"
	requisitionadapter "github.com/AppeiYA/requisition-system/internal/attachment/adapters/out/requisition"
	"github.com/AppeiYA/requisition-system/internal/attachment/adapters/out/storage/local"
	attachmenttelemetry "github.com/AppeiYA/requisition-system/internal/attachment/adapters/out/telemetry"
	attachmentdomain "github.com/AppeiYA/requisition-system/internal/attachment/domain"
	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
	attachmentusecase "github.com/AppeiYA/requisition-system/internal/attachment/usecase"
	identityhttp "github.com/AppeiYA/requisition-system/internal/identity/adapters/in/http"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/generator"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/mime"
	"github.com/gin-gonic/gin"
)

// Attachment test limits are deliberately small: the upload cap must be
// reachable without pushing megabytes through the multipart reader, and one
// attachment per parent makes the over-count rule observable.
const (
	AttachmentTestMaxSizeBytes = 2 << 10
	AttachmentTestMaxPerRef    = 1
	AttachmentTestSignedTTL    = 5 * time.Minute
	AttachmentTestStagingTTL   = time.Hour
)

type AttachmentHTTP struct {
	*RequisitionHTTP
	AppDB                  *db.DB
	Router                 *gin.Engine
	PortIn                 *attachmentin.AttachmentPortIn
	Reconcile              attachmentin.ReconcileOutboxUsecaseInt
	Cleanup                attachmentin.CleanupAttachmentsUsecaseInt
	FileStorage            attachmentout.FileStorage
	AttachmentRepository   *attachmentpostgres.AttachmentRepository
	StagedUploadRepository *attachmentpostgres.StagedUploadRepository
	OutboxRepository       *attachmentpostgres.ReconciliationOutboxRepository
	LocalRoot              string
	StagingTTL             time.Duration
	MaxSizeBytes           int64
	MaxPerRef              int
}

// requisitionAttachmentAccessChecker resolves a polymorphic attachment ref
// through the requisition port. This is the same delegation the composition
// root uses, kept local so the HTTP fixture has no dependency on app wiring.
type requisitionAttachmentAccessChecker struct {
	requisitionPortIn *requisitionin.RequisitionPortIn
}

func (c *requisitionAttachmentAccessChecker) CanAccess(
	ctx context.Context,
	refName, refID, actorID, actorRole string,
	_ attachmentout.Action,
) error {
	switch attachmentdomain.AttachmentRefName(refName) {
	case attachmentdomain.RefNameRequisition:
		_, err := c.requisitionPortIn.GetRequisitionUsecase.Execute(ctx, requisitionin.GetRequisitionQuery{
			RequisitionID: refID,
			ActorID:       actorID,
			ActorRole:     actorRole,
		})
		return err
	case attachmentdomain.RefNameRequisitionItem:
		_, err := c.requisitionPortIn.GetRequisitionItemUsecase.Execute(ctx, requisitionin.GetRequisitionItemQuery{
			RequisitionItemID: refID,
			ActorID:           actorID,
			ActorRole:         actorRole,
		})
		return err
	default:
		return attachmentdomain.ErrInvalidAttachmentRefName
	}
}

func (s *Suite) SetupAttachmentHTTP(t *testing.T) *AttachmentHTTP {
	t.Helper()
	return s.SetupAttachmentHTTPWithLimits(
		t,
		AttachmentTestMaxSizeBytes,
		AttachmentTestMaxPerRef,
	)
}

func (s *Suite) SetupAttachmentHTTPWithLimits(
	t *testing.T,
	maxSizeBytes int64,
	maxPerRef int,
) *AttachmentHTTP {
	t.Helper()

	localRoot := filepath.Join(t.TempDir(), "attachments")
	if err := os.MkdirAll(localRoot, 0o700); err != nil {
		t.Fatalf("create attachment storage root: %v", err)
	}

	baseRepository := db.NewRepository(s.AppDB)
	idGenerator := &generator.IDGenerator{}
	transactionManager := db.NewTransactionManager(s.AppDB)
	repository := attachmentpostgres.NewAttachmentRepository(baseRepository)
	stagedRepository := attachmentpostgres.NewStagedUploadRepository(baseRepository)
	outboxRepository := attachmentpostgres.NewReconciliationOutboxRepository(baseRepository)
	claim := attachmentusecase.NewClaimStagedAttachmentsUsecase(
		stagedRepository, repository, outboxRepository, idGenerator,
	)
	registrar := requisitionadapter.NewStagedAttachmentRegistrar(claim)
	requisition := s.SetupRequisitionHTTPWithRegistrar(t, registrar)

	fileStorage := local.New(localRoot)
	detector := mime.New()
	accessChecker := &requisitionAttachmentAccessChecker{
		requisitionPortIn: requisition.RequisitionPortIn,
	}

	var signer attachmentout.SignedURLProvider
	if provider, ok := any(fileStorage).(attachmentout.SignedURLProvider); ok {
		signer = provider
	}

	upload := attachmentusecase.NewUploadAttachmentUsecase(
		repository, fileStorage, detector, accessChecker, idGenerator, transactionManager,
		maxSizeBytes, maxPerRef,
	)
	stage := attachmentusecase.NewStageUploadUsecase(
		stagedRepository, fileStorage, detector, idGenerator, transactionManager,
		maxSizeBytes, AttachmentTestStagingTTL,
	)
	download := attachmentusecase.NewDownloadAttachmentUsecase(
		repository, fileStorage, accessChecker, signer, AttachmentTestSignedTTL,
	)
	get := attachmentusecase.NewGetAttachmentUsecase(repository, accessChecker)
	list := attachmentusecase.NewListAttachmentsUsecase(repository, accessChecker)
	deleteAttachment := attachmentusecase.NewDeleteAttachmentUsecase(
		repository, fileStorage, accessChecker, transactionManager,
	)
	deleteForRef := attachmentusecase.NewDeleteAttachmentsForRefUsecase(
		repository, fileStorage, transactionManager,
	)
	metrics, err := attachmenttelemetry.NewRecorder()
	if err != nil {
		t.Fatalf("create attachment metrics recorder: %v", err)
	}
	reconcile := attachmentusecase.NewReconcileOutboxUsecase(
		outboxRepository, repository, fileStorage, transactionManager, metrics,
		50, 5, time.Minute,
	)
	cleanup := attachmentusecase.NewCleanupAttachmentsUsecase(
		stagedRepository, repository, outboxRepository, fileStorage, transactionManager, metrics,
		50, 15*time.Minute,
	)
	portIn := attachmentin.NewAttachmentPortIn(
		upload, stage, download, get, list, deleteAttachment, deleteForRef,
	)

	// Reuse the requisition fixture's router so one HTTP client can create a
	// parent requisition and then attach files to it.
	gin.SetMode(gin.TestMode)
	router := requisition.Router
	authMiddleware := identityhttp.NewAuthMiddleware(
		requisition.AccessTokens,
		requisition.UserRepository,
	)
	attachmenthttp.NewAttachmentHandler(portIn, maxSizeBytes, "/api/v1").Register(
		router.Group("/api/v1"),
		authMiddleware.Handle(),
	)

	return &AttachmentHTTP{
		RequisitionHTTP:        requisition,
		AppDB:                  s.AppDB,
		Router:                 router,
		PortIn:                 portIn,
		Reconcile:              reconcile,
		Cleanup:                cleanup,
		FileStorage:            fileStorage,
		AttachmentRepository:   repository,
		StagedUploadRepository: stagedRepository,
		OutboxRepository:       outboxRepository,
		LocalRoot:              localRoot,
		StagingTTL:             AttachmentTestStagingTTL,
		MaxSizeBytes:           maxSizeBytes,
		MaxPerRef:              maxPerRef,
	}
}
