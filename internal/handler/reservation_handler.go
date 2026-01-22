package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (h *Handlers) CreateReservation(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "reservation created"})
}

func (h *Handlers) UpdateReservation(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "reservation updated"})
}

func (h *Handlers) GetReservation(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "get reservation success"})
}
func (h *Handlers) GetReservationByQuoteId(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "get reservation by quote ID success"})

}
func (h *Handlers) GetReservationByOrderId(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "get reservation by order ID success"})
}

func (h *Handlers) DeleteReservation(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "reservation deleted"})
}
func (h *Handlers) CommitReservation(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "reservation committed success"})
}
func (h *Handlers) Attach(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "reservation attached success"})
}
func (h *Handlers) GetReservationAvailability(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "reservation availability success"})
}
