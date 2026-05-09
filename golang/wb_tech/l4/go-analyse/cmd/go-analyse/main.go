package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/gin-gonic/gin"

	"go-analyse/internal/config"
	"go-analyse/internal/metrics"
	"go-analyse/internal/server"
)

func main() {
	cfg := config.LoadFromEnv()
	gin.SetMode(cfg.GinMode)

	previousGCPercent := debug.SetGCPercent(cfg.GCPercent)
	log.Printf("gc percent set to %d (previous %d)", cfg.GCPercent, previousGCPercent)

	registry := metrics.NewRegistry(cfg.GCPercent)
	handler := server.NewHandler(registry)

	httpServer := &http.Server{
		Addr:              cfg.Address(),
		Handler:           handler,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("starting go-analyse server on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server failed: %v", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Println("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
