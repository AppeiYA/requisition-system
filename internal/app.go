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
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	a.config = cfg

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	api := router.Group("/api")

	api.GET("/healthz", func(c *gin.Context) {
		response.Success(c, "OK", http.StatusOK, nil)
	})

	NewRouter(api)

	a.server = server.New(router, &cfg.Server)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCH := make(chan error, 1)
	go func() {
		log.Printf("Server is running on port: %s", cfg.Server.Address)
		if err := a.server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCH <- err
		}
	}()

	select {
	case err := <-errCH:
		return err
	case <-ctx.Done():
		log.Println("Shutting down server...")
	}

	shutDownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return a.server.Shutdown(shutDownCtx)
}
