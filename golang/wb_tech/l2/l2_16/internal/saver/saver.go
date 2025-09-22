package saver

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Saver struct {
	outputDir string
}

func NewSaver(outputDir string) (*Saver, error) {
	return &Saver{outputDir: outputDir}, nil
}

func (s *Saver) Save(u *url.URL, body []byte, contentType string) (string, error) {
	localPath := s.UrlToPath(u)
	if err := os.MkdirAll(filepath.Dir(localPath), os.ModePerm); err != nil {
		return "", err
	}

	return localPath, os.WriteFile(localPath, body, 0644)
}

func (s *Saver) UrlToPath(u *url.URL) string {
	path := filepath.Join(s.outputDir, u.Path)
	if strings.HasSuffix(u.Path, "/") || u.Path == "" {
		path = filepath.Join(path, "index.html")
	}
	return path
}
