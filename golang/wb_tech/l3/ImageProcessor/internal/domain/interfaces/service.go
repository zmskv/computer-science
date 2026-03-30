package interfaces

import (
	"context"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/domain/entity"
)

type ImageService interface {
	Upload(ctx context.Context, filename string, data []byte) (entity.Image, error)
	Get(ctx context.Context, id string) (entity.Image, error)
	List(ctx context.Context) ([]entity.Image, error)
	Delete(ctx context.Context, id string) error
	Process(ctx context.Context, id string) error
}
