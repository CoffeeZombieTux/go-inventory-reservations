package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestPingRoute verifies that the ping endpoint returns 200 with pong message.
func TestPingRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.Default()

	setupAppHealthRoutes(engine)

	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	resp := httptest.NewRecorder()

	engine.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.JSONEq(t, `{"success":true,"message":"pong"}`, resp.Body.String())
}
