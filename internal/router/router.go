package router

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(engine *gin.Engine) {
	setupAppHealthRoutes(engine)
}
