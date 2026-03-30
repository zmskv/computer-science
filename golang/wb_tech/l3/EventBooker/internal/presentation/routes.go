package presentation

import (
	"github.com/wb-go/wbf/ginext"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/EventBooker/internal/domain/interfaces"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/EventBooker/internal/presentation/handlers"
	"go.uber.org/zap"
)

func InitRoutes(
	r *ginext.Engine,
	service interfaces.EventService,
	logger *zap.Logger,
	webRoot string,
) {
	h := handlers.NewHandler(service, logger, webRoot)

	r.GET("/", h.Index)
	r.GET("/app.js", h.AppJS)
	r.GET("/styles.css", h.StylesCSS)

	r.GET("/users", h.ListUsers)
	r.POST("/users", h.CreateUser)
	r.GET("/events", h.ListEvents)
	r.POST("/events", h.CreateEvent)
	r.GET("/events/:id", h.GetEvent)
	r.POST("/events/:id/book", h.BookSeat)
	r.POST("/events/:id/confirm", h.ConfirmBooking)
}
