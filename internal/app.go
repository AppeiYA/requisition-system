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

	"github.com/AppeiYA/requisition-system/internal/shared/config"
	"github.com/AppeiYA/requisition-system/internal/shared/response"
	"github.com/AppeiYA/requisition-system/internal/shared/server"
	"github.com/AppeiYA/requisition-system/internal/shared/telemetry"
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

	otel:= telemetry.NewTelemetry("requisition-system")
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

	NewRouter(api)

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

	if err := a.server.Shutdown(ctx); err != nil {
		runErr = errors.Join(runErr, err)
	}

	if err := shutdownOtel(shutDownCtx); err != nil {
		runErr = errors.Join(runErr, err)
	}

	return runErr
}
