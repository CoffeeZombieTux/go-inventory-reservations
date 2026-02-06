package router

import (
	"github.com/gin-gonic/gin"
	"go-inventory-reservations/internal/handler"
)

func setupStockRoutes(engine *gin.Engine, handlersPool handler.HandlersPool) {
	stock := engine.Group("/stock")
	stock.GET("/:sku", handlersPool.Stock.GetStockBySku)
	stock.GET("/", handlersPool.Stock.GetStocks) // use limit and offset query params
}
