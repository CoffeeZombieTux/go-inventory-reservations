package router

import (
	"github.com/gin-gonic/gin"
	"go-inventory-reservations/internal/handler"
)

func setupStockRoutes(engine *gin.Engine, handlers handler.Handlers) {
	admin := engine.Group("/stock")
	admin.GET("/stock/:sku", handlers.GetStockBySku)
	admin.POST("/stock/:skus", handlers.GetStockBySkus)
}
