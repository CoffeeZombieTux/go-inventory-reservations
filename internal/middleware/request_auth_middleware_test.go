package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-inventory-reservations/internal/apperror"
	apimodel "go-inventory-reservations/internal/model/api"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestRequestID_UsesIncomingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(RequestID())
	engine.GET("/ping", func(ctx *gin.Context) {
		if got := requestIDFromContext(ctx); got != "req-abc" {
			t.Fatalf("expected request id req-abc, got %q", got)
		}
		ctx.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("X-Request-Id", "req-abc")
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", w.Code)
	}
	if got := w.Header().Get("X-Request-Id"); got != "req-abc" {
		t.Fatalf("expected response request id req-abc, got %q", got)
	}
}

func TestRequestID_GeneratesWhenMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(RequestID())
	engine.GET("/ping", func(ctx *gin.Context) {
		ctx.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	got := w.Header().Get("X-Request-Id")
	if got == "" {
		t.Fatalf("expected generated request id")
	}
	if _, err := uuid.Parse(got); err != nil {
		t.Fatalf("expected generated request id to be uuid, got %q", got)
	}
}

func TestBearerTokenAuth_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(RequestID())
	engine.GET("/protected", BearerTokenAuth("secret"), func(ctx *gin.Context) {
		ctx.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d", w.Code)
	}
	var resp apimodel.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Error == nil || resp.Error.Code != apperror.CodeUnauthorizedCode {
		t.Fatalf("unexpected error payload: %+v", resp.Error)
	}
	if resp.Error.RequestID == "" {
		t.Fatalf("expected request_id to be set")
	}
}

func TestBearerTokenAuth_InvalidPrefix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(RequestID())
	engine.GET("/protected", BearerTokenAuth("secret"), func(ctx *gin.Context) {
		ctx.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic abc")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d", w.Code)
	}
}

func TestBearerTokenAuth_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(RequestID())
	engine.GET("/protected", BearerTokenAuth("secret"), func(ctx *gin.Context) {
		ctx.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer secret")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected success, got %d", w.Code)
	}
}
