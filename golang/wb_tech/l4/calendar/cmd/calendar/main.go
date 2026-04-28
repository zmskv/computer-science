package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"calendar/internal/di"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	container := di.New(ctx, os.Stdout, di.LoadConfigFromEnv())
	defer container.Close()

	go func() {
		container.Logger.Info("starting server", map[string]any{"addr": container.Server.Addr})
		if err := container.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			container.Logger.Error("server failed", map[string]any{"error": err.Error()})
		}
	}()

	<-ctx.Done()
	container.Logger.Info("shutting down server", nil)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), container.Config.ShutdownTimeout)
	defer cancel()

	if err := container.Server.Shutdown(shutdownCtx); err != nil {
		container.Logger.Error("server forced to shutdown", map[string]any{"error": err.Error()})
	}

	container.Logger.Info("server exiting", nil)
}
