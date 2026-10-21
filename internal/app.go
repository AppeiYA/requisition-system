package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	jwtadapter "github.com/AppeiYA/requisition-system/internal/identity/adapters/out/jwt"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/generator"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/redis"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/server"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/taskqueue"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/telemetry"
	"github.com/gin-gonic/gin"
)

type App struct {
	config *config.Config
	server *server.Server
}

func NewApp() *App {

	return &App{}
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	a.config = cfg

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

	otel := telemetry.NewTelemetry("requisition-system")
	shutdownOtel, err := otel.Setup(ctx)
	if err != nil {
		return err
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	api := router.Group("/api")

	api.GET("/healthz", func(c *gin.Context) {
		response.Success(c, "OK", http.StatusOK, nil)
	})

	txManager := db.NewTransactionManager(database)
	accessTokenService := jwtadapter.NewJWTService(cfg.JWT)
	idGen := &generator.IDGenerator{}

	identityModule := a.buildIdentity(cfg, database, redisClient, txManager, accessTokenService, idGen)

	departmentModule := a.buildDepartment(database, idGen)

	invitationModule := a.buildInvitation(
		database,
		txManager,
		identityModule.PortIn().RegisterUserUseCase,
		departmentModule.PortIn().CreateMembershipUsecase,
		idGen,
	)

	// The asynq client shares the Redis pool above; redisClient.Close()
	// owns tearing it down, so the client is not closed separately.
	asynqClient := taskqueue.NewClient(redisClient)

	adminModule := a.buildAdminHandler(
		cfg,
		txManager,
		identityModule.PortIn().GetUserByEmailUsecase,
		invitationModule.PortIn().SaveInvitationUsecase,
		departmentModule.PortIn().GetDepartmentByIDUsecase,
		asynqClient)

	if err := NewRouter(
		api,
		identityModule.Handler(),
		invitationModule.Handler(),
		adminModule.Handler(),
		departmentModule.Handler(),
		identityModule.AuthMiddleware(),
	); err != nil {
		return err
	}

	a.server = server.New(router, &cfg.Server)

	errCH := make(chan error, 1)
	go func() {
		log.Printf("Server is running on port: %s", cfg.Server.Address)
		if err := a.server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCH <- err
		}
	}()

	var runErr error

	select {
	case err := <-errCH:
		runErr = err
	case <-ctx.Done():
		log.Println("Shutting down server...")
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
