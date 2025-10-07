package presentation

import (
	"github.com/wb-go/wbf/ginext"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/DelayedNotifier/internal/domain/interfaces"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/DelayedNotifier/internal/presentation/handlers"
	"go.uber.org/zap"
)

func InitRoutes(r *ginext.Engine, notifierService interfaces.NotifierService, logger *zap.Logger) {
	notifierHandler := handlers.NewHandler(notifierService, logger)

	notifier := r.Group("/notifier")
	{
		notifier.POST("/notify", notifierHandler.Notify)
		notifier.GET("/notify/:id", notifierHandler.GetNotifyById)
		notifier.DELETE("/notify/:id", notifierHandler.DeleteNotifyById)
	}

}
