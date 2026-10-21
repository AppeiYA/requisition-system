package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	adminTasks "github.com/AppeiYA/requisition-system/internal/admin/tasks"
	invitationworker "github.com/AppeiYA/requisition-system/internal/invitation/adapters/in/worker"
	invitationpostgres "github.com/AppeiYA/requisition-system/internal/invitation/adapters/out/postgres"
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

	otel := telemetry.NewTelemetry("requisition-system-worker")
	shutdownOtel, err := otel.Setup(ctx)
	if err != nil {
		return err
	}

	mux := asynq.NewServeMux()

	adminModule := w.buildAdminWorker(cfg)
	mux.HandleFunc(adminTasks.SendInvitationEmail, adminModule.SendMailHandler().Handle)

	invitationRepo := invitationpostgres.NewInvitationRepository(db.NewRepository(database))
	expireInvitations := invitationusecase.NewExpireInvitationsUsecase(invitationRepo)
	expireInvitationsHandler := invitationworker.NewExpireInvitationsHandler(expireInvitations)
	mux.HandleFunc(invitationworker.ExpireInvitations, expireInvitationsHandler.Handle)

	w.server = taskqueue.NewServer(redisClient, cfg.Worker)
	scheduler := taskqueue.NewScheduler(redisClient)
	if _, err := scheduler.Register("0 0 * * *", asynq.NewTask(invitationworker.ExpireInvitations, nil), asynq.Queue(cfg.Worker.Queue)); err != nil {
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
