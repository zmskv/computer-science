package ginapp

import (
	"github.com/gin-gonic/gin"

	"calendar/internal/domain/interfaces"
	"calendar/internal/presentation/http/ginapp/handlers"
)

func InitRoutes(r *gin.Engine, eventService interfaces.EventService, logger interfaces.Logger) {
	eventHandler := handlers.NewHandler(eventService, logger)

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	calendar := r.Group("/calendar")
	{
		calendar.POST("/create_event", eventHandler.CreateEvent)
		calendar.POST("/update_event", eventHandler.UpdateEvent)
		calendar.POST("/delete_event", eventHandler.DeleteEvent)
		calendar.GET("/events_for_day", eventHandler.EventsForDay)
		calendar.GET("/events_for_week", eventHandler.EventsForWeek)
		calendar.GET("/events_for_month", eventHandler.EventsForMonth)
		calendar.GET("/archived_events", eventHandler.ArchivedEvents)
	}
}
