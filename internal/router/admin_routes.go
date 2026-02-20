package router

import (
	"github.com/gin-gonic/gin"
	"go-inventory-reservations/internal/handler"
)

// setupAdminRoutes sets up the admin routes
func setupAdminRoutes(admin *gin.RouterGroup, handlersPool handler.HandlersPool) {
	admin.POST("/stock", handlersPool.Admin.CreateStock)
	admin.PUT("/stock", handlersPool.Admin.UpdateStock)
	admin.DELETE("/stock/:sku", handlersPool.Admin.DeleteStock)
	admin.GET("/stock/:sku/reservation-items", handlersPool.Admin.GetActiveReservationItemsBySku)
}
