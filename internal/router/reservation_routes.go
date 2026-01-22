package router

import (
	"github.com/gin-gonic/gin"
	"go-inventory-reservations/internal/handler"
)

func setupReservationRoutes(engine *gin.Engine, handlers handler.Handlers) {
	admin := engine.Group("/reservation")
	admin.POST("/reservation", handlers.CreateReservation)
	admin.PUT("/reservation", handlers.UpdateReservation)
	admin.GET("/reservation/:id", handlers.GetReservation)
	admin.GET("/reservation/quote/:quote_id", handlers.GetReservationByQuoteId)
	admin.GET("/reservation/order/:order_id", handlers.GetReservationByOrderId)
	admin.DELETE("/reservation/:id", handlers.DeleteReservation)
	admin.GET("/reservation/:id/commit", handlers.CommitReservation)
	admin.POST("/reservation/:id/attach", handlers.Attach)
	admin.GET("/reservation/:id/availability", handlers.GetReservationAvailability)

}
