package presentation

import (
	"github.com/wb-go/wbf/ginext"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/domain/interfaces"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/CommentTree/internal/presentation/handlers"
	"go.uber.org/zap"
)

func InitRoutes(r *ginext.Engine, commentService interfaces.CommentService, logger *zap.Logger) {
	r.Use(corsMiddleware())

	commentHandler := handlers.NewHandler(commentService, logger)

	r.POST("/comments", commentHandler.CreateComment)
	r.GET("/comments", commentHandler.GetComments)
	r.DELETE("/comments/:id", commentHandler.DeleteComment)
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
