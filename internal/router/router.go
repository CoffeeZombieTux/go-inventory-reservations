package router

import (
	"github.com/gin-gonic/gin"
	"go-inventory-reservations/internal/handler"
)

func SetupRoutes(engine *gin.Engine, handlersPool handler.HandlersPool) {
	setupAppHealthRoutes(engine)
	setupAdminRoutes(engine, handlersPool)
	setupStockRoutes(engine, handlersPool)
	setupReservationRoutes(engine, handlersPool)
}
