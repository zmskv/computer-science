package ginapp

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/zmskv/computer-science/golang/wb_tech/l2/l2_18/internal/domain/interfaces"
	"github.com/zmskv/computer-science/golang/wb_tech/l2/l2_18/internal/presentation/http/ginapp/handlers"

	"go.uber.org/zap"
)

func InitRoutes(r *gin.Engine, eventService interfaces.EventService, logger *zap.Logger) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	eventHandler := handlers.NewHandler(eventService, logger)

	calendar := r.Group("/calendar")
	{
		calendar.POST("/create_event", eventHandler.CreateEvent)
		calendar.POST("/update_event", eventHandler.UpdateEvent)
		calendar.POST("/delete_event", eventHandler.DeleteEvent)
		calendar.GET("/events_for_day", eventHandler.EventsForDay)
		calendar.GET("/events_for_week", eventHandler.EventsForWeek)
		calendar.GET("/events_for_month", eventHandler.EventsForMonth)
	}
}
