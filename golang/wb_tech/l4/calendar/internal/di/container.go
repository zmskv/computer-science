package di

import (
	"context"
	"io"
	"net/http"
	"os"
	"time"

	"calendar/internal/application"
	"calendar/internal/infrastructure/logging"
	"calendar/internal/infrastructure/reminder"
	"calendar/internal/infrastructure/repository"
	"calendar/internal/presentation/http/ginapp"
	"calendar/internal/presentation/http/ginapp/middleware"

	"github.com/gin-gonic/gin"
)

type Config struct {
	Port            string
	GinMode         string
	LogBuffer       int
	ReminderBuffer  int
	ArchiveInterval time.Duration
	ArchiveAfter    time.Duration
	ShutdownTimeout time.Duration
}

type Container struct {
	Config Config
	Logger *logging.AsyncLogger
	Server *http.Server
}

func DefaultConfig() Config {
	return Config{
		Port:            "8080",
		GinMode:         gin.ReleaseMode,
		LogBuffer:       256,
		ReminderBuffer:  128,
		ArchiveInterval: time.Minute,
		ArchiveAfter:    24 * time.Hour,
		ShutdownTimeout: 10 * time.Second,
	}
}

func LoadConfigFromEnv() Config {
	cfg := DefaultConfig()

	if port := os.Getenv("PORT"); port != "" {
		cfg.Port = port
	}
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		cfg.GinMode = mode
	}
	cfg.ArchiveInterval = envDuration("ARCHIVE_INTERVAL", cfg.ArchiveInterval)
	cfg.ArchiveAfter = envDuration("ARCHIVE_AFTER", cfg.ArchiveAfter)
	cfg.ShutdownTimeout = envDuration("SHUTDOWN_TIMEOUT", cfg.ShutdownTimeout)

	return cfg
}

func New(ctx context.Context, writer io.Writer, cfg Config) *Container {
	cfg = withDefaults(cfg)

	logger := logging.NewAsyncLogger(writer, cfg.LogBuffer)
	repo := repository.NewInMemoryRepo()
	service := application.NewService(
		repo,
		reminder.NewLogNotifier(logger),
		logger,
		application.Config{
			ReminderBuffer:  cfg.ReminderBuffer,
			ArchiveInterval: cfg.ArchiveInterval,
			ArchiveAfter:    cfg.ArchiveAfter,
		},
	)
	service.Start(ctx)

	gin.SetMode(cfg.GinMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.LoggerMiddleware(logger))
	ginapp.InitRoutes(router, service, logger)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	return &Container{
		Config: cfg,
		Logger: logger,
		Server: server,
	}
}

func (c *Container) Close() {
	if c == nil || c.Logger == nil {
		return
	}
	c.Logger.Close()
}

func withDefaults(cfg Config) Config {
	defaults := DefaultConfig()

	if cfg.Port == "" {
		cfg.Port = defaults.Port
	}
	if cfg.GinMode == "" {
		cfg.GinMode = defaults.GinMode
	}
	if cfg.LogBuffer <= 0 {
		cfg.LogBuffer = defaults.LogBuffer
	}
	if cfg.ReminderBuffer <= 0 {
		cfg.ReminderBuffer = defaults.ReminderBuffer
	}
	if cfg.ArchiveInterval <= 0 {
		cfg.ArchiveInterval = defaults.ArchiveInterval
	}
	if cfg.ArchiveAfter <= 0 {
		cfg.ArchiveAfter = defaults.ArchiveAfter
	}
	if cfg.ShutdownTimeout <= 0 {
		cfg.ShutdownTimeout = defaults.ShutdownTimeout
	}

	return cfg
}

func envDuration(name string, fallback time.Duration) time.Duration {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return value
}
