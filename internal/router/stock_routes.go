package router

import (
	"go-inventory-reservations/internal/handler"

	"github.com/gin-gonic/gin"
)

// setupStockRoutes sets up the stock routes
func setupStockRoutes(engine *gin.Engine, handlersPool handler.HandlersPool, publicAuth gin.HandlerFunc) {
	stock := engine.Group("/stock")
	stock.Use(publicAuth)
	stock.GET("/:sku", handlersPool.Stock.GetStockBySku)
	stock.GET("", handlersPool.Stock.GetStocks) // use limit and offset query params
}
