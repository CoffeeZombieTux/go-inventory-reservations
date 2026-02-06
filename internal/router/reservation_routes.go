package router

import (
	"github.com/gin-gonic/gin"
	"go-inventory-reservations/internal/handler"
)

func setupReservationRoutes(engine *gin.Engine, handlersPool handler.HandlersPool) {
	reservation := engine.Group("/reservation")
	reservation.POST("", handlersPool.Reservation.CreateReservation)
	reservation.PUT("", handlersPool.Reservation.UpdateReservation)
	reservation.GET("/:id", handlersPool.Reservation.GetReservationById)
	reservation.GET("/quote/:quote_id", handlersPool.Reservation.GetReservationByQuoteId)
	reservation.GET("/order/:order_id", handlersPool.Reservation.GetReservationByOrderId)
	reservation.DELETE("/:id", handlersPool.Reservation.DeleteReservation)
	reservation.GET("/:id/commit", handlersPool.Reservation.CommitReservation)
	reservation.POST("/:id/attach", handlersPool.Reservation.Attach)
	reservation.GET("/:id/availability", handlersPool.Reservation.GetReservationAvailability)
}
