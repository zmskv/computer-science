package application

import (
	"context"
	"errors"
	"os"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/application/dto"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/domain/interfaces"
	"go.uber.org/zap"
)

type ImageWorker struct {
	service  interfaces.ImageService
	consumer interfaces.ImageJobConsumer
	logger   *zap.Logger
}

func NewImageWorker(service interfaces.ImageService, consumer interfaces.ImageJobConsumer, logger *zap.Logger) *ImageWorker {
	return &ImageWorker{
		service:  service,
		consumer: consumer,
		logger:   logger,
	}
}

func (w *ImageWorker) Run(ctx context.Context) error {
	w.logger.Info("Image worker started")

	err := w.consumer.Consume(ctx, func(ctx context.Context, job dto.ImageJob) error {
		if job.ImageID == "" {
			w.logger.Warn("skipping empty image job")
			return nil
		}

		w.logger.Info("processing image job", zap.String("image_id", job.ImageID))

		if err := w.service.Process(ctx, job.ImageID); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				w.logger.Warn("image not found for job", zap.String("image_id", job.ImageID))
				return nil
			}

			w.logger.Error("image processing failed", zap.String("image_id", job.ImageID), zap.Error(err))
			return err
		}

		w.logger.Info("image job completed", zap.String("image_id", job.ImageID))
		return nil
	})

	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	w.logger.Info("Image worker stopped")
	return nil
}
