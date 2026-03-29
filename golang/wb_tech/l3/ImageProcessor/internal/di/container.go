package di

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/wb-go/wbf/ginext"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/application"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/domain/interfaces"
	kafkaadapter "github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/infrastructure/messaging/kafka"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/infrastructure/processing"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/infrastructure/repository/metadata"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/infrastructure/storage/filesystem"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/presentation"
	projectlogger "github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/logger"
	"go.uber.org/zap"
)

type Container struct {
	Config     *Config
	Logger     *zap.Logger
	HTTPServer *http.Server
}

func NewContainer(ctx context.Context) *Container {
	cfg := ReadConfig()
	log := projectlogger.New()

	storage, err := filesystem.NewStorage(cfg.Storage.RootDir)
	if err != nil {
		log.Fatal("failed to initialize image storage", zap.Error(err))
	}

	repo, err := metadata.NewRepository(filepath.Join(cfg.Storage.RootDir, "metadata"))
	if err != nil {
		log.Fatal("failed to initialize metadata repository", zap.Error(err))
	}

	if err := kafkaadapter.EnsureTopic(ctx, cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.Partitions, cfg.Kafka.ReplicationFactor, cfg.Kafka.RetryStrategy()); err != nil {
		log.Fatal("failed to ensure kafka topic", zap.String("topic", cfg.Kafka.Topic), zap.Error(err))
	}

	processor := processing.NewProcessor(log)
	publisher := kafkaadapter.NewPublisher(cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.RetryStrategy())
	consumer := kafkaadapter.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.Group, cfg.Kafka.RetryStrategy(), log)

	imageService := application.NewImageService(
		repo,
		storage,
		processor,
		publisher,
		cfg.Processing.MaxUploadSizeBytes(),
		cfg.Processing.Options(),
		log,
	)

	worker := application.NewImageWorker(imageService, consumer, log)
	go func() {
		if err := worker.Run(ctx); err != nil {
			log.Error("image worker stopped with error", zap.Error(err))
		}
	}()

	httpServer := InitHTTPServer(cfg.HTTP, imageService, log, cfg.Storage.RootDir, cfg.Web.Dir)

	return &Container{
		Config:     &cfg,
		Logger:     log,
		HTTPServer: httpServer,
	}
}

func InitHTTPServer(
	cfg HTTPConfig,
	imageService interfaces.ImageService,
	logger *zap.Logger,
	storageRoot string,
	webDir string,
) *http.Server {
	router := ginext.New("")
	presentation.InitRoutes(router, imageService, logger, storageRoot, webDir)

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
	if c.HTTPServer == nil {
		return nil
	}

	return c.HTTPServer.Shutdown(ctx)
}
