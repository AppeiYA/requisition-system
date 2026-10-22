package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
	adminTasks "github.com/AppeiYA/requisition-system/internal/admin/tasks"
	invitationworker "github.com/AppeiYA/requisition-system/internal/invitation/adapters/in/worker"
	invitationpostgres "github.com/AppeiYA/requisition-system/internal/invitation/adapters/out/postgres"
	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
	invitationusecase "github.com/AppeiYA/requisition-system/internal/invitation/usecase"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/redis"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/taskqueue"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/telemetry"
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
			log.Printf("failed to flush telemetry: %v", err)
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
			log.Printf("failed to close database: %v", err)
		}
	}()

	redisClient, err := redis.Connect(cfg.Redis)
	if err != nil {
		return err
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("failed to close redis client: %v", err)
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
	)
	if err != nil {
		return err
	}

	mux := asynq.NewServeMux()
	// The tracing middleware wraps every handler, so the whole of a task is
	// traced, including the database and email calls of the use cases below.
	mux.Use(taskqueue.TraceMiddleware())

	invitationRepo := invitationpostgres.NewInvitationRepository(db.NewRepository(database))
	adminModule := w.buildAdminWorker(cfg, invitationRepo, operationTelemetry)
	mux.HandleFunc(adminTasks.SendInvitationEmail, adminModule.SendMailHandler().Handle)

	expireInvitations := invitationusecase.NewExpireInvitationsUsecase(invitationRepo, operationTelemetry)
	expireInvitationsHandler := invitationworker.NewExpireInvitationsHandler(expireInvitations)
	mux.HandleFunc(invitationworker.ExpireInvitations, expireInvitationsHandler.Handle)

	w.server = taskqueue.NewServer(redisClient, cfg.Worker)
	scheduler := taskqueue.NewScheduler(redisClient)
	if _, err := scheduler.Register("*/15 * * * *", asynq.NewTask(invitationworker.ExpireInvitations, nil), asynq.Queue(cfg.Worker.Queue)); err != nil {
		return err
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
	log.Printf("Worker started: queue=%s concurrency=%d", cfg.Worker.Queue, cfg.Worker.Concurrency)

	<-ctx.Done()
	log.Println("Shutting down worker...")
	w.server.Shutdown() // lets in-flight tasks finish within ShutdownTimeout

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := shutdownOtel(shutdownCtx); err != nil {
		return err
	}

	return nil
}
