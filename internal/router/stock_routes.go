package router

import (
	"github.com/gin-gonic/gin"
	"go-inventory-reservations/internal/handler"
)

// setupStockRoutes sets up the stock routes
func setupStockRoutes(engine *gin.Engine, handlersPool handler.HandlersPool) {
	stock := engine.Group("/stock")
	stock.GET("/:sku", handlersPool.Stock.GetStockBySku)
	stock.GET("/", handlersPool.Stock.GetStocks) // use limit and offset query params
}
