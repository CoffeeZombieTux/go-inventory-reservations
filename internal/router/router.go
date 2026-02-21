package router

import (
	"github.com/gin-gonic/gin"
	"go-inventory-reservations/internal/config"
	"go-inventory-reservations/internal/handler"
	"go-inventory-reservations/internal/logger"
	"go-inventory-reservations/internal/middleware"
)

// SetupRoutes sets up the application routes
func SetupRoutes(engine *gin.Engine, handlersPool handler.HandlersPool, cfg *config.Config, log *logger.Logger) {
	engine.Use(middleware.RequestID())
	engine.Use(middleware.HTTPLogging(log))

	setupAppHealthRoutes(engine)

	admin := engine.Group("/admin")
	admin.Use(middleware.BearerTokenAuth(cfg.Auth.AdminToken))
	setupAdminRoutes(admin, handlersPool)

	public := middleware.BearerTokenAuth(cfg.Auth.PublicToken)
	setupStockRoutes(engine, handlersPool, public)
	setupReservationRoutes(engine, handlersPool, public)
}
