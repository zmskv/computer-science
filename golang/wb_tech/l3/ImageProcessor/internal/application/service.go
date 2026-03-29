package application

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/application/dto"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/domain/entity"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/domain/interfaces"
	"go.uber.org/zap"
)

var errUnsupportedImage = errors.New("supported formats are jpg, png, gif")

type imageService struct {
	images             interfaces.ImageRepository
	storage            interfaces.ImageStorage
	processor          interfaces.ImageProcessor
	publisher          interfaces.ImageJobPublisher
	maxUploadSizeBytes int64
	options            entity.ProcessingOptions
	logger             *zap.Logger
}

func NewImageService(
	images interfaces.ImageRepository,
	storage interfaces.ImageStorage,
	processor interfaces.ImageProcessor,
	publisher interfaces.ImageJobPublisher,
	maxUploadSizeBytes int64,
	options entity.ProcessingOptions,
	logger *zap.Logger,
) interfaces.ImageService {
	return &imageService{
		images:             images,
		storage:            storage,
		processor:          processor,
		publisher:          publisher,
		maxUploadSizeBytes: maxUploadSizeBytes,
		options:            options,
		logger:             logger,
	}
}

func (s *imageService) Upload(ctx context.Context, input dto.UploadImageInput) (entity.Image, error) {
	if len(input.Data) == 0 {
		return entity.Image{}, errors.New("image file is empty")
	}

	if s.maxUploadSizeBytes > 0 && int64(len(input.Data)) > s.maxUploadSizeBytes {
		return entity.Image{}, fmt.Errorf("image exceeds %d MB limit", s.maxUploadSizeBytes/(1024*1024))
	}

	format, err := detectImageFormat(input.Data)
	if err != nil {
		return entity.Image{}, err
	}

	now := time.Now().UTC()
	imageID := uuid.NewString()

	originalPath, err := s.storage.SaveOriginal(ctx, imageID, format, input.Data)
	if err != nil {
		return entity.Image{}, fmt.Errorf("save original image: %w", err)
	}

	imageMeta := entity.Image{
		ID:               imageID,
		OriginalFilename: filepath.Base(input.Filename),
		Format:           format,
		Status:           entity.StatusQueued,
		OriginalPath:     originalPath,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.images.Save(ctx, imageMeta); err != nil {
		_ = s.storage.Delete(ctx, originalPath)
		return entity.Image{}, fmt.Errorf("save image metadata: %w", err)
	}

	if err := s.publisher.Publish(ctx, dto.ImageJob{ImageID: imageID}); err != nil {
		_ = s.storage.Delete(ctx, originalPath)
		_ = s.images.Delete(ctx, imageID)
		return entity.Image{}, fmt.Errorf("enqueue image processing: %w", err)
	}

	return imageMeta, nil
}

func (s *imageService) Get(ctx context.Context, id string) (entity.Image, error) {
	return s.images.GetByID(ctx, id)
}

func (s *imageService) List(ctx context.Context) ([]entity.Image, error) {
	return s.images.List(ctx)
}

func (s *imageService) Delete(ctx context.Context, id string) error {
	imageMeta, err := s.images.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.storage.Delete(ctx, imageMeta.OriginalPath, imageMeta.ProcessedPath, imageMeta.ThumbnailPath); err != nil {
		return fmt.Errorf("delete image files: %w", err)
	}

	if err := s.images.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete image metadata: %w", err)
	}

	return nil
}

func (s *imageService) Process(ctx context.Context, id string) error {
	imageMeta, err := s.images.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if imageMeta.Status == entity.StatusReady && imageMeta.ProcessedPath != "" && imageMeta.ThumbnailPath != "" {
		return nil
	}

	imageMeta.Status = entity.StatusProcessing
	imageMeta.Error = ""
	imageMeta.UpdatedAt = time.Now().UTC()

	if err := s.images.Save(ctx, imageMeta); err != nil {
		return fmt.Errorf("update image status to processing: %w", err)
	}

	originalData, err := s.storage.Read(ctx, imageMeta.OriginalPath)
	if err != nil {
		return s.markFailed(ctx, imageMeta, fmt.Errorf("read original image: %w", err))
	}

	processed, err := s.processor.Process(ctx, originalData, imageMeta.Format, s.options)
	if err != nil {
		return s.markFailed(ctx, imageMeta, fmt.Errorf("process image: %w", err))
	}

	processedPath, err := s.storage.SaveProcessed(ctx, imageMeta.ID, processed.Format, processed.Processed)
	if err != nil {
		return s.markFailed(ctx, imageMeta, fmt.Errorf("save processed image: %w", err))
	}

	thumbnailPath, err := s.storage.SaveThumbnail(ctx, imageMeta.ID, processed.Format, processed.Thumbnail)
	if err != nil {
		_ = s.storage.Delete(ctx, processedPath)
		return s.markFailed(ctx, imageMeta, fmt.Errorf("save thumbnail image: %w", err))
	}

	imageMeta.Status = entity.StatusReady
	imageMeta.ProcessedPath = processedPath
	imageMeta.ThumbnailPath = thumbnailPath
	imageMeta.Error = ""
	imageMeta.UpdatedAt = time.Now().UTC()

	if err := s.images.Save(ctx, imageMeta); err != nil {
		return fmt.Errorf("save processed metadata: %w", err)
	}

	return nil
}

func (s *imageService) markFailed(ctx context.Context, imageMeta entity.Image, cause error) error {
	imageMeta.Status = entity.StatusFailed
	imageMeta.Error = cause.Error()
	imageMeta.UpdatedAt = time.Now().UTC()

	if err := s.images.Save(ctx, imageMeta); err != nil {
		s.logger.Error("failed to persist image error state", zap.String("image_id", imageMeta.ID), zap.Error(err))
	}

	return cause
}

func detectImageFormat(data []byte) (string, error) {
	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("decode image config: %w", err)
	}

	switch format {
	case "jpeg", "jpg":
		return "jpg", nil
	case "png":
		return "png", nil
	case "gif":
		return "gif", nil
	default:
		return "", errUnsupportedImage
	}
}
