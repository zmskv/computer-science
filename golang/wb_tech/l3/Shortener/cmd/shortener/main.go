package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/Shortener/internal/di"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	container := di.NewContainer(ctx)
	logger := container.Logger
	defer logger.Sync()

	go func() {
		logger.Info("HTTP server is starting", zap.String("addr", container.HTTPServer.Addr))
		if err := container.HTTPServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("HTTP server failed to start", zap.Error(err))
		}
	}()

	<-ctx.Done()

	waitForShutdown(cancel, container)
}

func waitForShutdown(cancel context.CancelFunc, container *di.Container) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	container.Logger.Info("Shutting down...")

	cancel()

	if err := container.HTTPServer.Shutdown(context.Background()); err != nil {
		container.Logger.Error("order service stop failed", zap.Error(err))
	}
}
