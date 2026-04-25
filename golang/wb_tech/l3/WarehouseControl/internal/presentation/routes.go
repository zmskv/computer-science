package presentation

import (
	"github.com/wb-go/wbf/ginext"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/internal/domain/interfaces"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/internal/presentation/handlers"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/internal/presentation/middleware"
	"go.uber.org/zap"
)

func InitRoutes(
	r *ginext.Engine,
	warehouseService interfaces.WarehouseService,
	authService interfaces.AuthService,
	logger *zap.Logger,
	webRoot string,
) {
	r.Use(corsMiddleware())

	h := handlers.NewHandler(warehouseService, authService, logger, webRoot)

	r.GET("/", h.Index)
	r.GET("/app.js", h.AppJS)
	r.GET("/styles.css", h.StylesCSS)
	r.GET("/health", h.Health)

	r.POST("/auth/login", h.Login)

	authRequired := middleware.Auth(authService)

	r.GET("/items", authRequired, h.ListItems)
	r.GET("/items/:id", authRequired, h.GetItem)
	r.POST("/items", authRequired, h.CreateItem)
	r.PUT("/items/:id", authRequired, h.UpdateItem)
	r.DELETE("/items/:id", authRequired, h.DeleteItem)

	r.GET("/items/:id/history", authRequired, h.ListItemHistory)
	r.GET("/history", authRequired, h.ListHistory)
	r.GET("/history/export", authRequired, h.ExportHistoryCSV)
}

func corsMiddleware() ginext.HandlerFunc {
	return func(c *ginext.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
