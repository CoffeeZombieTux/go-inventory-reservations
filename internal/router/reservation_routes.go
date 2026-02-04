package router

import (
	"github.com/gin-gonic/gin"
	"go-inventory-reservations/internal/handler"
)

func setupReservationRoutes(engine *gin.Engine, handlersPool handler.HandlersPool) {
	admin := engine.Group("/reservation")
	admin.POST("/reservation", handlersPool.Reservation.CreateReservation)
	admin.PUT("/reservation", handlersPool.Reservation.UpdateReservation)
	admin.GET("/reservation/:id", handlersPool.Reservation.GetReservation)
	admin.GET("/reservation/quote/:quote_id", handlersPool.Reservation.GetReservationByQuoteId)
	admin.GET("/reservation/order/:order_id", handlersPool.Reservation.GetReservationByOrderId)
	admin.DELETE("/reservation/:id", handlersPool.Reservation.DeleteReservation)
	admin.GET("/reservation/:id/commit", handlersPool.Reservation.CommitReservation)
	admin.POST("/reservation/:id/attach", handlersPool.Reservation.Attach)
	admin.GET("/reservation/:id/availability", handlersPool.Reservation.GetReservationAvailability)

}
