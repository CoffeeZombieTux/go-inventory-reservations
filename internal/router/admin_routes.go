package router

import (
	"github.com/gin-gonic/gin"
	"go-inventory-reservations/internal/handler"
)

func setupAdminRoutes(engine *gin.Engine, handlers handler.Handlers) {
	admin := engine.Group("/admin")
	admin.POST("/stock", handlers.CreateStock)
	admin.PUT("/stock", handlers.UpdateStock)
	admin.DELETE("/stock", handlers.DeleteStock)
}
