package handlers

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/wb-go/wbf/ginext"
	appdto "github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/application/dto"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/domain/entity"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/domain/interfaces"
	presentationdto "github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/presentation/dto"
	"go.uber.org/zap"
)

type Handler struct {
	service     interfaces.ImageService
	logger      *zap.Logger
	storageRoot string
	webRoot     string
}

func NewHandler(service interfaces.ImageService, logger *zap.Logger, storageRoot, webRoot string) *Handler {
	return &Handler{
		service:     service,
		logger:      logger,
		storageRoot: storageRoot,
		webRoot:     webRoot,
	}
}

func (h *Handler) Index(c *ginext.Context) {
	h.serveFrontendFile(c, "index.html", "text/html; charset=utf-8")
}

func (h *Handler) AppJS(c *ginext.Context) {
	h.serveFrontendFile(c, "app.js", "application/javascript; charset=utf-8")
}

func (h *Handler) StylesCSS(c *ginext.Context) {
	h.serveFrontendFile(c, "styles.css", "text/css; charset=utf-8")
}

func (h *Handler) Upload(c *ginext.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		h.logger.Error("failed to read multipart form file", zap.Error(err))
		c.JSON(http.StatusBadRequest, ginext.H{"error": "file field is required"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		h.logger.Error("failed to read uploaded file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "failed to read file"})
		return
	}

	imageMeta, err := h.service.Upload(c.Request.Context(), appdto.UploadImageInput{
		Filename: header.Filename,
		Data:     data,
	})
	if err != nil {
		h.logger.Error("failed to upload image", zap.Error(err))
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, presentationdto.FromEntity(imageMeta))
}

func (h *Handler) ListImages(c *ginext.Context) {
	images, err := h.service.List(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to list images", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "failed to list images"})
		return
	}

	response := make([]presentationdto.ImageResponse, 0, len(images))
	for _, imageMeta := range images {
		response = append(response, presentationdto.FromEntity(imageMeta))
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetImage(c *ginext.Context) {
	id := c.Param("id")

	imageMeta, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		h.respondNotFound(c, err, "image not found")
		return
	}

	if c.Query("meta") == "1" {
		c.JSON(http.StatusOK, presentationdto.FromEntity(imageMeta))
		return
	}

	h.serveProcessedFile(c, imageMeta)
}

func (h *Handler) DeleteImage(c *ginext.Context) {
	id := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		h.respondNotFound(c, err, "image not found")
		return
	}

	c.JSON(http.StatusOK, ginext.H{"status": "deleted"})
}

func (h *Handler) ServeStoredFile(c *ginext.Context) {
	fullPath, err := h.resolveStoragePath(c.Param("filepath"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid file path"})
		return
	}

	if _, err := os.Stat(fullPath); err != nil {
		h.respondNotFound(c, err, "file not found")
		return
	}

	http.ServeFile(c.Writer, c.Request, fullPath)
}

func (h *Handler) serveProcessedFile(c *ginext.Context, imageMeta entity.Image) {
	if imageMeta.Status != entity.StatusReady || imageMeta.ProcessedPath == "" {
		c.JSON(http.StatusConflict, ginext.H{"error": "image is still processing"})
		return
	}

	fullPath, err := h.resolveStoragePath("/" + imageMeta.ProcessedPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid file path"})
		return
	}

	if _, err := os.Stat(fullPath); err != nil {
		h.respondNotFound(c, err, "processed file not found")
		return
	}

	http.ServeFile(c.Writer, c.Request, fullPath)
}

func (h *Handler) serveFrontendFile(c *ginext.Context, filename, contentType string) {
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Header("Content-Type", contentType)
	c.File(filepath.Join(h.webRoot, filename))
}

func (h *Handler) respondNotFound(c *ginext.Context, err error, message string) {
	if errors.Is(err, os.ErrNotExist) {
		c.JSON(http.StatusNotFound, ginext.H{"error": message})
		return
	}

	h.logger.Error(message, zap.Error(err))
	c.JSON(http.StatusInternalServerError, ginext.H{"error": message})
}

func (h *Handler) resolveStoragePath(requestPath string) (string, error) {
	cleanPath := filepath.Clean(filepath.FromSlash(strings.TrimPrefix(requestPath, "/")))
	if cleanPath == "." || strings.HasPrefix(cleanPath, "..") {
		return "", errors.New("invalid path")
	}

	rootAbs, err := filepath.Abs(h.storageRoot)
	if err != nil {
		return "", err
	}

	fullPath := filepath.Join(rootAbs, cleanPath)
	fullAbs, err := filepath.Abs(fullPath)
	if err != nil {
		return "", err
	}

	rel, err := filepath.Rel(rootAbs, fullAbs)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(rel, "..") {
		return "", errors.New("invalid path")
	}

	return fullAbs, nil
}
