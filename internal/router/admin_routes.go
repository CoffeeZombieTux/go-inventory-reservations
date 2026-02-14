package router

import (
	"github.com/gin-gonic/gin"
	"go-inventory-reservations/internal/handler"
)

// setupAdminRoutes sets up the admin routes
func setupAdminRoutes(engine *gin.Engine, handlersPool handler.HandlersPool) {
	admin := engine.Group("/admin")
	admin.POST("/stock", handlersPool.Admin.CreateStock)
	admin.PUT("/stock", handlersPool.Admin.UpdateStock)
	admin.DELETE("/stock/:sku", handlersPool.Admin.DeleteStock)
}
