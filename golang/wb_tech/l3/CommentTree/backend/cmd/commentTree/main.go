package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/di"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	container := di.NewContainer(ctx)

	go func() {
		container.Logger.Info("Starting HTTP server", zap.String("addr", container.HTTPServer.Addr))
		if err := container.HTTPServer.ListenAndServe(); err != nil {
			container.Logger.Fatal("HTTP server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	container.Logger.Info("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), container.Config.HTTP.ShutdownTimeout)
	defer cancel()

	if err := container.Shutdown(shutdownCtx); err != nil {
		container.Logger.Error("Error during shutdown", zap.Error(err))
	}
}
