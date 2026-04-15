package presentation

import (
	"github.com/wb-go/wbf/ginext"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/SalesTracker/internal/domain/interfaces"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/SalesTracker/internal/presentation/handlers"
	"go.uber.org/zap"
)

func InitRoutes(
	r *ginext.Engine,
	service interfaces.SalesService,
	logger *zap.Logger,
	webRoot string,
) {
	h := handlers.NewHandler(service, logger, webRoot)

	r.GET("/", h.Index)
	r.GET("/app.js", h.AppJS)
	r.GET("/styles.css", h.StylesCSS)
	r.GET("/health", h.Health)

	r.POST("/items", h.CreateItem)
	r.GET("/items", h.ListItems)
	r.GET("/items/export", h.ExportItemsCSV)
	r.GET("/items/:id", h.GetItem)
	r.PUT("/items/:id", h.UpdateItem)
	r.DELETE("/items/:id", h.DeleteItem)

	r.GET("/analytics", h.GetAnalytics)
}
