package di

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/wb-go/wbf/ginext"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/Shortener/internal/application"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/Shortener/internal/domain/interfaces"
	conn "github.com/zmskv/computer-science/golang/wb_tech/l3/Shortener/internal/infrastructure/db/postgres"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/Shortener/internal/infrastructure/repository/postgres"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/Shortener/logger"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/Shortener/internal/presentation"
	"go.uber.org/zap"
)

type Container struct {
	Config     *Config
	HTTPServer *http.Server
	Logger     *zap.Logger
}

func NewContainer(ctx context.Context) *Container {
	cfg := ReadConfig()
	log := logger.New()

	masterDSN := cfg.Postgres.BuildMasterDSN()
	if masterDSN == "" {
		log.Fatal("PG master DSN is required")
	}

	db, err := conn.NewDB(ctx, masterDSN, cfg.Postgres.SlaveDSNs)
	if err != nil {
		log.Fatal("failed to init postgres repo", zap.Error(err))
	}
	urlRepo := postgres.NewURLRepository(db, log)
	analyticsRepo := postgres.NewAnalyticsRepository(db, log)

	service := application.NewShortenerService(urlRepo, analyticsRepo)

	httpServer := InitHTTPServer(cfg.HTTP, service, log)

	return &Container{
		Config:     &cfg,
		HTTPServer: httpServer,
		Logger:     log,
	}
}

func InitHTTPServer(cfg HTTPConfig, notifierService interfaces.ShortenerService, logger *zap.Logger) *http.Server {
	router := ginext.New("")
	presentation.InitRoutes(router, notifierService, logger)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	return &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
}

func (c *Container) Shutdown(ctx context.Context) error {
	if c.HTTPServer != nil {
		return c.HTTPServer.Shutdown(ctx)
	}
	return nil
}
