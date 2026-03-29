package filesystem

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	dirOriginals  = "originals"
	dirProcessed  = "processed"
	dirThumbnails = "thumbnails"
)

type Storage struct {
	rootDir string
}

func NewStorage(rootDir string) (*Storage, error) {
	dirs := []string{
		rootDir,
		filepath.Join(rootDir, dirOriginals),
		filepath.Join(rootDir, dirProcessed),
		filepath.Join(rootDir, dirThumbnails),
		filepath.Join(rootDir, "metadata"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	return &Storage{rootDir: rootDir}, nil
}

func (s *Storage) SaveOriginal(ctx context.Context, imageID, format string, data []byte) (string, error) {
	return s.save(ctx, dirOriginals, imageID, format, data)
}

func (s *Storage) SaveProcessed(ctx context.Context, imageID, format string, data []byte) (string, error) {
	return s.save(ctx, dirProcessed, imageID, format, data)
}

func (s *Storage) SaveThumbnail(ctx context.Context, imageID, format string, data []byte) (string, error) {
	return s.save(ctx, dirThumbnails, imageID, format, data)
}

func (s *Storage) Read(ctx context.Context, path string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	fullPath := filepath.Join(s.rootDir, filepath.FromSlash(path))
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", fullPath, err)
	}

	return data, nil
}

func (s *Storage) Delete(ctx context.Context, paths ...string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	for _, relPath := range paths {
		if relPath == "" {
			continue
		}

		fullPath := filepath.Join(s.rootDir, filepath.FromSlash(relPath))
		if err := os.Remove(fullPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("delete file %s: %w", fullPath, err)
		}
	}

	return nil
}

func (s *Storage) save(ctx context.Context, dir, imageID, format string, data []byte) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	filename := imageID + "." + format
	fullPath := filepath.Join(s.rootDir, dir, filename)
	if err := os.WriteFile(fullPath, data, 0o644); err != nil {
		return "", fmt.Errorf("write file %s: %w", fullPath, err)
	}

	return filepath.ToSlash(filepath.Join(dir, filename)), nil
}
