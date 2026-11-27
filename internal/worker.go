package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
	adminTasks "github.com/AppeiYA/requisition-system/internal/admin/tasks"
	analyticsworker "github.com/AppeiYA/requisition-system/internal/analytics/adapters/in/worker"
	analyticspostgres "github.com/AppeiYA/requisition-system/internal/analytics/adapters/out/postgres"
	attachmentworker "github.com/AppeiYA/requisition-system/internal/attachment/adapters/in/worker"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	identitytasks "github.com/AppeiYA/requisition-system/internal/identity/tasks"
	invitationworker "github.com/AppeiYA/requisition-system/internal/invitation/adapters/in/worker"
	invitationpostgres "github.com/AppeiYA/requisition-system/internal/invitation/adapters/out/postgres"
	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
	invitationusecase "github.com/AppeiYA/requisition-system/internal/invitation/usecase"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/email"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/logger"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/redis"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/taskqueue"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/telemetry"
	"github.com/AppeiYA/requisition-system/internal/version"
	"github.com/hibiken/asynq"
)

// Worker runs the background task processor. It is the worker-process
// counterpart of App: same composition-root role, asynq instead of gin.
type Worker struct {
	config *config.Config
	server *asynq.Server
}

func NewWorker() *Worker {
	return &Worker{}
}

func (w *Worker) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	w.config = cfg

	otel := telemetry.NewTelemetry("requisition-system-worker", cfg.Telemetry)
	workerLogger := logger.New(logger.Options{
		Level:          cfg.Log.Level,
		Format:         cfg.Log.Format,
		AddSource:      cfg.Log.AddSource,
		ServiceName:    otel.ServiceName(),
		ServiceVersion: version.Version,
		Environment:    string(cfg.App.Env),
	})
	logger.SetDefault(workerLogger)

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
			workerLogger.Error("failed to flush telemetry", slog.String("error.message", err.Error()))
		}
	}()

	// The handlers are registered after the clients, so the database and
	// the queue calls of a task are traced from the very first one.
	database, err := db.Connect(cfg.DB)
	if err != nil {
		return err
	}
	defer func() {
		if err := database.Close(); err != nil {
			workerLogger.Error("failed to close database", slog.String("error.message", err.Error()))
		}
	}()

	redisClient, err := redis.Connect(cfg.Redis)
	if err != nil {
		return err
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			workerLogger.Error("failed to close redis client", slog.String("error.message", err.Error()))
		}
	}()

	// The operation names and the metric bases are stable and low cardinality:
	// no invitation, email or user identifier is ever reported.
	operationTelemetry, err := telemetry.NewOperationTelemetry(
		telemetry.Operation{
			Name:       adminout.OperationSendInvitationEmail,
			MetricBase: "requisition.admin.email_delivery",
			Unit:       "{email}",
		},
		telemetry.Operation{
			Name:       invitationout.OperationExpireInvitations,
			MetricBase: "requisition.invitation.expirations",
			Unit:       "{invitation}",
		},
		telemetry.Operation{
			Name:       identityout.OperationSendPasswordResetEmail,
			MetricBase: "requisition.auth.password_reset_emails",
			Unit:       "{email}",
		},
		telemetry.Operation{
			Name:       attachmentout.OperationReconcile,
			MetricBase: "requisition.attachment.reconciliations",
			Unit:       "{attachment}",
		},
		telemetry.Operation{
			Name:       attachmentout.OperationCleanup,
			MetricBase: "requisition.attachment.cleanups",
			Unit:       "{sweep}",
		},
	)
	if err != nil {
		return err
	}

	mux := asynq.NewServeMux()
	// The tracing and logging middlewares wrap every handler, so the whole of a task is
	// traced and logged with duration, task metadata, and trace context.
	mux.Use(taskqueue.TraceMiddleware(), taskqueue.LoggingMiddleware())

	invitationRepo := invitationpostgres.NewInvitationRepository(db.NewRepository(database))
	// One sender instance is shared by both worker modules so they do not
	// create separate instrumented HTTP clients.
	emailSender := email.NewSender(cfg.Email)
	adminModule := w.buildAdminWorker(cfg, invitationRepo, emailSender, operationTelemetry)
	mux.HandleFunc(adminTasks.SendInvitationEmail, adminModule.SendMailHandler().Handle)

	identityModule := w.buildIdentityWorker(cfg, database, emailSender, operationTelemetry)
	mux.HandleFunc(identitytasks.SendPasswordResetEmail, identityModule.SendPasswordResetEmailHandler().Handle)
	mux.HandleFunc(identitytasks.CleanupPasswordResetTokens, identityModule.CleanupPasswordResetTokensHandler().Handle)

	expireInvitations := invitationusecase.NewExpireInvitationsUsecase(invitationRepo, operationTelemetry)
	expireInvitationsHandler := invitationworker.NewExpireInvitationsHandler(expireInvitations)
	mux.HandleFunc(invitationworker.ExpireInvitations, expireInvitationsHandler.Handle)

	attachmentRepos := newAttachmentRepositories(db.NewRepository(database))
	attachmentModule, err := buildAttachmentWorker(
		cfg, attachmentRepos, db.NewTransactionManager(database), operationTelemetry,
	)
	if err != nil {
		return err
	}
	mux.HandleFunc(
		attachmentworker.ReconcileOutbox,
		attachmentModule.ReconcileOutboxHandler().Handle,
	)
	mux.HandleFunc(
		attachmentworker.Cleanup,
		attachmentModule.CleanupHandler().Handle,
	)

	if cfg.App.AnalyticsSource == "materialized" {
		analyticsRefresher := analyticspostgres.NewRefresher(db.NewRepository(database))
		analyticsRefreshHandler := analyticsworker.NewRefreshHandler(analyticsRefresher, redisClient.Client())
		mux.HandleFunc(analyticsworker.RefreshMaterializedViews, analyticsRefreshHandler.Handle)
	}

	w.server = taskqueue.NewServer(redisClient, cfg.Worker)
	scheduler := taskqueue.NewScheduler(redisClient)
	if _, err := scheduler.Register("*/15 * * * *", asynq.NewTask(invitationworker.ExpireInvitations, nil), asynq.Queue(cfg.Worker.Queue)); err != nil {
		return err
	}
	if _, err := scheduler.Register("0 * * * *", asynq.NewTask(identitytasks.CleanupPasswordResetTokens, nil), asynq.Queue(cfg.Worker.Queue)); err != nil {
		return err
	}
	if _, err := scheduler.Register(
		cfg.Attachment.ReconcileSchedule,
		asynq.NewTask(attachmentworker.ReconcileOutbox, nil),
		asynq.Queue(cfg.Worker.Queue),
	); err != nil {
		return err
	}
	if _, err := scheduler.Register(
		cfg.Attachment.CleanupSchedule,
		asynq.NewTask(attachmentworker.Cleanup, nil),
		asynq.Queue(cfg.Worker.Queue),
	); err != nil {
		return err
	}
	if cfg.App.AnalyticsSource == "materialized" {
		if _, err := scheduler.Register("*/5 * * * *", asynq.NewTask(analyticsworker.RefreshMaterializedViews, nil), asynq.Queue(cfg.Worker.Queue)); err != nil {
			return err
		}
	}
	if err := scheduler.Start(); err != nil {
		return err
	}
	defer scheduler.Shutdown()

	// Start instead of Run: Run traps SIGTERM/SIGINT itself, and the worker
	// shares App's signal handling so both binaries shut down the same way.
	if err := w.server.Start(mux); err != nil {
		return err
	}
	workerLogger.Info("worker started", slog.String("queue", cfg.Worker.Queue), slog.Int("concurrency", cfg.Worker.Concurrency))

	<-ctx.Done()
	workerLogger.Info("shutting down worker")
	w.server.Shutdown() // lets in-flight tasks finish within ShutdownTimeout

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := shutdownOtel(shutdownCtx); err != nil {
		return err
	}

	return nil
}
