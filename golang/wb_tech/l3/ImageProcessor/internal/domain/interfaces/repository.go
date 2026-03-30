package interfaces

import (
	"context"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/domain/entity"
)

type ImageRepository interface {
	Save(ctx context.Context, image entity.Image) error
	GetByID(ctx context.Context, id string) (entity.Image, error)
	List(ctx context.Context) ([]entity.Image, error)
	Delete(ctx context.Context, id string) error
}

type ImageStorage interface {
	SaveOriginal(ctx context.Context, imageID, format string, data []byte) (string, error)
	SaveProcessed(ctx context.Context, imageID, format string, data []byte) (string, error)
	SaveThumbnail(ctx context.Context, imageID, format string, data []byte) (string, error)
	Read(ctx context.Context, path string) ([]byte, error)
	Delete(ctx context.Context, paths ...string) error
}

type ImageProcessor interface {
	Process(ctx context.Context, source []byte, format string, options entity.ProcessingOptions) (entity.ProcessedImage, error)
}

type ImageJobPublisher interface {
	Publish(ctx context.Context, imageID string) error
}

type ImageJobConsumer interface {
	Consume(ctx context.Context, handler func(context.Context, string) error) error
}
