package presentation

import (
	"github.com/wb-go/wbf/ginext"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/domain/interfaces"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/presentation/handlers"
	"go.uber.org/zap"
)

func InitRoutes(r *ginext.Engine, commentService interfaces.CommentService, logger *zap.Logger) {
	commentHandler := handlers.NewHandler(commentService, logger)

	r.POST("/comments", commentHandler.CreateComment)
	r.GET("/comments", commentHandler.GetComments)
	r.DELETE("/comments/:id", commentHandler.DeleteComment)
}
