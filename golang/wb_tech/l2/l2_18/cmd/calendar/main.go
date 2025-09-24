package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zmskv/computer-science/golang/wb_tech/l2/l2_18/internal/application"
	"github.com/zmskv/computer-science/golang/wb_tech/l2/l2_18/internal/infrastructure/repository"
	"github.com/zmskv/computer-science/golang/wb_tech/l2/l2_18/internal/presentation/http/ginapp"
	"github.com/zmskv/computer-science/golang/wb_tech/l2/l2_18/internal/presentation/http/ginapp/middleware"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	repo := repository.NewInMemoryRepo()
	service := application.NewService(repo)

	r := gin.New()
	r.Use(middleware.LoggerMiddleware(logger))

	ginapp.InitRoutes(r, service, logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exiting")
}
