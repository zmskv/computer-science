package presentation

import (
	"github.com/wb-go/wbf/ginext"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/Shortener/internal/domain/interfaces"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/Shortener/internal/presentation/handlers"
	"go.uber.org/zap"
)

func InitRoutes(r *ginext.Engine, service interfaces.ShortenerService, logger *zap.Logger) {
	h := handlers.NewHandler(service, logger)

	r.POST("/shorten", h.Shorten)
	r.GET("/s/:short_url", h.Redirect)
	r.GET("/analytics/:short_url", h.Analytics)

}
