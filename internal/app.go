package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
	attachmentexternal "github.com/AppeiYA/requisition-system/internal/attachment/adapters/out/external"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
	jwtadapter "github.com/AppeiYA/requisition-system/internal/identity/adapters/out/jwt"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/generator"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/logger"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/redis"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/server"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/taskqueue"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/telemetry"
	"github.com/AppeiYA/requisition-system/internal/version"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type App struct {
	config *config.Config
	server *server.Server
}

func NewApp() *App {

	return &App{}
}

func (a *App) Config() (*config.Config, error) {
	if a.config == nil {
		cfg, err := config.Load()
		if err != nil {
			return nil, err
		}
		a.config = cfg
	}
	return a.config, nil
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	a.config = cfg

	otel := telemetry.NewTelemetry("requisition-system", cfg.Telemetry)
	appLogger := logger.New(logger.Options{
		Level:          cfg.Log.Level,
		Format:         cfg.Log.Format,
		AddSource:      cfg.Log.AddSource,
		ServiceName:    otel.ServiceName(),
		ServiceVersion: version.Version,
		Environment:    string(cfg.App.Env),
	})
	logger.SetDefault(appLogger)

	shutdownOtel, err := otel.Setup(ctx)
	if err != nil {
		return err
	}
	// The instrumented clients below capture the providers installed above,
	// so the telemetry has to be set up before them. That also means a
	// startup failure below returns with spans still buffered, hence this
	// deferred flush. Shutting the providers down twice is harmless: it only
	// runs once, and the explicit call at the end of the run reports the
	// error.
	defer func() {
		if err := shutdownOtel(context.Background()); err != nil {
			appLogger.Error("failed to flush telemetry", slog.String("error.message", err.Error()))
		}
	}()

	database, err := db.Connect(cfg.DB)
	if err != nil {
		return err
	}
	defer func() {
		if err := database.Close(); err != nil {
			appLogger.Error("failed to close database", slog.String("error.message", err.Error()))
		}
	}()

	redisClient, err := redis.Connect(cfg.Redis)
	if err != nil {
		return err
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			appLogger.Error("failed to close redis client", slog.String("error.message", err.Error()))
		}
	}()

	// The operation names and the metric bases are stable and low cardinality:
	// no invitation, email or user identifier is ever reported.
	operationTelemetry, err := telemetry.NewOperationTelemetry(
		telemetry.Operation{
			Name:       adminout.OperationInviteMember,
			MetricBase: "requisition.admin.invitations",
			Unit:       "{invitation}",
		},
		telemetry.Operation{
			Name:       invitationout.OperationAcceptInvitation,
			MetricBase: "requisition.invitee.acceptances",
			Unit:       "{acceptance}",
		},
		telemetry.Operation{
			Name:       adminout.OperationUpdateUserStatus,
			MetricBase: "requisition.admin.user_status_updates",
			Unit:       "{update}",
		},
		telemetry.Operation{
			Name:       adminout.OperationUpdateUserRole,
			MetricBase: "requisition.admin.user_role_updates",
			Unit:       "{update}",
		},
		telemetry.Operation{
			Name:       adminout.OperationRevokeUserSessions,
			MetricBase: "requisition.admin.session_revocations",
			Unit:       "{revocation}",
		},
		telemetry.Operation{
			Name:       adminout.OperationResendInvitation,
			MetricBase: "requisition.admin.invitation_resends",
			Unit:       "{resend}",
		},
		telemetry.Operation{
			Name:       adminout.OperationRevokeInvitation,
			MetricBase: "requisition.admin.invitation_revocations",
			Unit:       "{revocation}",
		},
		telemetry.Operation{
			Name:       identityout.OperationRequestPasswordReset,
			MetricBase: "requisition.auth.password_reset_requests",
			Unit:       "{request}",
		},
		telemetry.Operation{
			Name:       identityout.OperationResetPassword,
			MetricBase: "requisition.auth.password_reset_completions",
			Unit:       "{completion}",
		},
		telemetry.Operation{
			Name:       attachmentout.OperationUpload,
			MetricBase: "requisition.attachment.uploads",
			Unit:       "{attachment}",
		},
		telemetry.Operation{
			Name:       attachmentout.OperationDownload,
			MetricBase: "requisition.attachment.downloads",
			Unit:       "{attachment}",
		},
		telemetry.Operation{
			Name:       attachmentout.OperationDelete,
			MetricBase: "requisition.attachment.deletions",
			Unit:       "{attachment}",
		},
	)
	if err != nil {
		return err
	}

	router := gin.New()
	if err := router.SetTrustedProxies(cfg.RateLimit.TrustedProxies); err != nil {
		return err
	}
	// The tracing middleware goes first, so the span of an operation covers
	// the whole request, including the work of the other middlewares. It
	// rewrites the request with the span context, so every handler picks it
	// up through c.Request.Context() without any change.
	router.Use(otelgin.Middleware(otel.ServiceName()))
	router.Use(logger.GinLogger(appLogger), logger.GinRecovery(appLogger))
	router.Use(requestBodyLimit(cfg.API.BaseURL))
	api := router.Group("/api")

	api.GET("/healthz", func(c *gin.Context) {
		response.Success(c, "OK", http.StatusOK, nil)
	})

	txManager := db.NewTransactionManager(database)
	baseRepo := db.NewRepository(database)
	accessTokenService := jwtadapter.NewJWTService(cfg.JWT)
	idGen := &generator.IDGenerator{}
	redisCache := redis.NewCache(redisClient.Client())

	// The asynq client is needed by the identity queue adapter, so it is
	// created before the identity module is composed.
	asynqClient := taskqueue.NewClient(redisClient)

	identityModule := a.buildIdentity(
		cfg,
		baseRepo,
		redisCache,
		redisClient,
		txManager,
		accessTokenService,
		idGen,
		asynqClient,
		operationTelemetry,
	)

	departmentModule := a.buildDepartment(baseRepo, idGen, redisCache)

	attachmentRepos := newAttachmentRepositories(baseRepo)
	stagedRegistrar := a.buildAttachmentRegistrar(attachmentRepos, idGen, cfg.Attachment.MaxPerRef)

	requisitionModule := a.buildRequisition(
		baseRepo, idGen, redisCache, departmentModule.MembershipRepository(),
		txManager, identityModule.UserLookup(), stagedRegistrar,
	)

	attachmentModule, err := a.buildAttachment(
		cfg,
		idGen,
		attachmentexternal.NewRequisitionAccessChecker(requisitionModule.RequisitionPortIn()),
		txManager,
		attachmentRepos,
	)
	if err != nil {
		return err
	}

	invitationModule := a.buildInvitation(
		baseRepo,
		txManager,
		identityModule.PortIn().RegisterUserUseCase,
		departmentModule.PortIn().CreateMembershipUsecase,
		idGen,
		operationTelemetry,
	)

	analyticsModule := buildAnalytics(baseRepo, redisCache, cfg.App.AnalyticsSource)

	adminModule := a.buildAdminHandler(
		cfg,
		txManager,
		identityModule.PortIn().GetUserByEmailUsecase,
		identityModule.AdminUsers(),
		invitationModule.PortIn().SaveInvitationUsecase,
		invitationModule.AdminInvitations(),
		departmentModule.PortIn().GetDepartmentByIDUsecase,
		asynqClient,
		operationTelemetry)

	if err := NewRouter(
		api,
		identityModule.Handler(),
		invitationModule.Handler(),
		adminModule.Handler(),
		departmentModule.Handler(),
		requisitionModule.ItemCategoryHandler(),
		requisitionModule.ItemHandler(),
		requisitionModule.RequisitionHandler(),
		attachmentModule.Handler(),
		analyticsModule.Handler(),
		identityModule.AuthMiddleware(),
	); err != nil {
		return err
	}

	a.server = server.New(router, &cfg.Server)

	errCH := make(chan error, 1)
	go func() {
		appLogger.Info("server is running", slog.String("server.address", cfg.Server.Address))
		if err := a.server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCH <- err
		}
	}()

	var runErr error

	select {
	case err := <-errCH:
		runErr = err
	case <-ctx.Done():
		appLogger.Info("shutting down server")
	}

	shutDownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.server.Shutdown(shutDownCtx); err != nil {
		runErr = errors.Join(runErr, err)
	}

	if err := shutdownOtel(shutDownCtx); err != nil {
		runErr = errors.Join(runErr, err)
	}

	return runErr
}

// buildIdentityHandler wires the identity use cases to their adapters. It is
// the only place that knows the concrete implementations, keeping the inner
// layers free of infrastructure dependencies.
