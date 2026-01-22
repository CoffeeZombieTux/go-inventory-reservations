package router

import (
	"github.com/gin-gonic/gin"
	"go-inventory-reservations/internal/handler"
)

func SetupRoutes(engine *gin.Engine, handlers handler.Handlers) {
	setupAppHealthRoutes(engine)
	setupAdminRoutes(engine, handlers)
	setupStockRoutes(engine, handlers)
	setupReservationRoutes(engine, handlers)
}
