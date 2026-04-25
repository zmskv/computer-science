package di

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/wb-go/wbf/ginext"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/internal/application"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/internal/domain/interfaces"
	postgresrepo "github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/internal/infrastructure/repository/postgres"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/internal/presentation"
	projectlogger "github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/logger"
	"go.uber.org/zap"
)

type Container struct {
	Config     *Config
	Logger     *zap.Logger
	HTTPServer *http.Server
	closers    []func()
}

func NewContainer(ctx context.Context) *Container {
	cfg := ReadConfig()
	log := projectlogger.New()

	repo, err := postgresrepo.NewRepository(ctx, cfg.Database.MasterDSN(), cfg.Database.SlaveDSNs, cfg.Database.Options())
	if err != nil {
		log.Fatal("failed to initialize repository", zap.Error(err))
	}

	warehouseService := application.NewWarehouseService(repo)
	authService := application.NewAuthService(cfg.Auth.JWTSecret, cfg.Auth.TokenTTL())
	httpServer := initHTTPServer(cfg.HTTP, warehouseService, authService, log, cfg.Web.Dir)

	return &Container{
		Config:     &cfg,
		Logger:     log,
		HTTPServer: httpServer,
		closers: []func(){
			repo.Close,
		},
	}
}

func initHTTPServer(
	cfg HTTPConfig,
	warehouseService interfaces.WarehouseService,
	authService interfaces.AuthService,
	logger *zap.Logger,
	webDir string,
) *http.Server {
	router := ginext.New("")
	presentation.InitRoutes(router, warehouseService, authService, logger, webDir)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	return &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func (c *Container) Shutdown(ctx context.Context) error {
	if c.HTTPServer != nil {
		if err := c.HTTPServer.Shutdown(ctx); err != nil {
			return err
		}
	}

	for _, closer := range c.closers {
		closer()
	}

	return nil
}
