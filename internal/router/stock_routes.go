package router

import (
	"github.com/gin-gonic/gin"
	"go-inventory-reservations/internal/handler"
)

func setupStockRoutes(engine *gin.Engine, handlersPool handler.HandlersPool) {
	admin := engine.Group("/stock")
	admin.GET("/:sku", handlersPool.Stock.GetStockBySku)
	admin.GET("/", handlersPool.Stock.GetStocks) // use limit and offset query params
}
