package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type AdminHandler struct {
}

func (h *Handlers) CreateStock(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "stock record created"})
}
func (h *Handlers) UpdateStock(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "stock record updated"})
}
func (h *Handlers) DeleteStock(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "stock record deleted"})

}
