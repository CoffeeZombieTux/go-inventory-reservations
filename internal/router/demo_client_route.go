package router

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed demo-client.html
var demoClientHTML []byte

func setupDemoClientRoute(engine *gin.Engine) {
	engine.GET("/demo", func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", demoClientHTML)
	})
}
