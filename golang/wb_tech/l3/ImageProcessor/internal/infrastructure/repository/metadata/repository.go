package metadata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/domain/entity"
)

type Repository struct {
	dir string
	mu  sync.RWMutex
}

func NewRepository(dir string) (*Repository, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create metadata directory: %w", err)
	}

	return &Repository{dir: dir}, nil
}

func (r *Repository) Save(ctx context.Context, imageMeta entity.Image) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(imageMeta, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	target := r.filePath(imageMeta.ID)
	temp := target + ".tmp"

	r.mu.Lock()
	defer r.mu.Unlock()

	if err := os.WriteFile(temp, data, 0o644); err != nil {
		return fmt.Errorf("write temp metadata file: %w", err)
	}

	if err := os.Rename(temp, target); err != nil {
		return fmt.Errorf("rename metadata file: %w", err)
	}

	return nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (entity.Image, error) {
	if err := ctx.Err(); err != nil {
		return entity.Image{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := os.ReadFile(r.filePath(id))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return entity.Image{}, fmt.Errorf("%w: image metadata %s", os.ErrNotExist, id)
		}
		return entity.Image{}, fmt.Errorf("read metadata: %w", err)
	}

	var imageMeta entity.Image
	if err := json.Unmarshal(data, &imageMeta); err != nil {
		return entity.Image{}, fmt.Errorf("unmarshal metadata: %w", err)
	}

	return imageMeta, nil
}

func (r *Repository) List(ctx context.Context) ([]entity.Image, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	files, err := filepath.Glob(filepath.Join(r.dir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("list metadata files: %w", err)
	}

	images := make([]entity.Image, 0, len(files))
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read metadata file: %w", err)
		}

		var imageMeta entity.Image
		if err := json.Unmarshal(data, &imageMeta); err != nil {
			return nil, fmt.Errorf("unmarshal metadata file: %w", err)
		}

		images = append(images, imageMeta)
	}

	sort.Slice(images, func(i, j int) bool {
		return images[i].CreatedAt.After(images[j].CreatedAt)
	})

	return images, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if err := os.Remove(r.filePath(id)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("delete metadata file: %w", err)
	}

	return nil
}

func (r *Repository) filePath(id string) string {
	return filepath.Join(r.dir, id+".json")
}
