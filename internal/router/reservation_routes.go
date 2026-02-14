package router

import (
	"github.com/gin-gonic/gin"
	"go-inventory-reservations/internal/handler"
)

// setupReservationRoutes sets up the reservation routes
func setupReservationRoutes(engine *gin.Engine, handlersPool handler.HandlersPool) {
	reservation := engine.Group("/reservation")
	reservation.POST("", handlersPool.Reservation.CreateReservation)
	reservation.PUT("", handlersPool.Reservation.UpdateReservation)
	reservation.GET("/:id", handlersPool.Reservation.GetReservationById)
	reservation.GET("/quote/:quote_id", handlersPool.Reservation.GetReservationByQuoteId)
	reservation.GET("/order/:order_id", handlersPool.Reservation.GetReservationByOrderId)
	reservation.POST("/attach-order", handlersPool.Reservation.AttachOrder)
	reservation.POST("/commit", handlersPool.Reservation.CommitReservation)
	reservation.GET("/:id/release", handlersPool.Reservation.ReleaseReservation)
	reservation.POST("/revert", handlersPool.Reservation.Revert)
}
