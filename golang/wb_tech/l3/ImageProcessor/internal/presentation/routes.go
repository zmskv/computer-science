package presentation

import (
	"github.com/wb-go/wbf/ginext"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/domain/interfaces"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/presentation/handlers"
	"go.uber.org/zap"
)

func InitRoutes(
	r *ginext.Engine,
	service interfaces.ImageService,
	logger *zap.Logger,
	storageRoot string,
	webRoot string,
) {
	h := handlers.NewHandler(service, logger, storageRoot, webRoot)

	r.GET("/", h.Index)
	r.GET("/app.js", h.AppJS)
	r.GET("/styles.css", h.StylesCSS)

	r.POST("/upload", h.Upload)
	r.GET("/images", h.ListImages)
	r.GET("/image/:id", h.GetImage)
	r.DELETE("/image/:id", h.DeleteImage)
	r.GET("/files/*filepath", h.ServeStoredFile)
}
