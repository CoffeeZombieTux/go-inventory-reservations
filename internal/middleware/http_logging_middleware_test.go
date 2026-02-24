package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNormalizeBodyForLog_NonTextContentType(t *testing.T) {
	text, truncated := normalizeBodyForLog("application/octet-stream", []byte{0x00, 0x01})
	if text != "<omitted: non-text body>" {
		t.Fatalf("unexpected text: %q", text)
	}
	if truncated {
		t.Fatalf("expected non-truncated flag")
	}
}

func TestNormalizeBodyForLog_TruncatesLongText(t *testing.T) {
	body := strings.Repeat("x", maxLoggedBodyBytes+100)
	text, truncated := normalizeBodyForLog("application/json", []byte(body))
	if len(text) != maxLoggedBodyBytes {
		t.Fatalf("expected truncated body length %d, got %d", maxLoggedBodyBytes, len(text))
	}
	if !truncated {
		t.Fatalf("expected truncated=true")
	}
}

func TestIsTextLikeContentType(t *testing.T) {
	cases := []struct {
		ct   string
		want bool
	}{
		{"", true},
		{"application/json", true},
		{"text/plain", true},
		{"application/octet-stream", false},
	}
	for _, tc := range cases {
		if got := isTextLikeContentType(tc.ct); got != tc.want {
			t.Fatalf("content type %q: want %v got %v", tc.ct, tc.want, got)
		}
	}
}

func TestBuildRequestLogFields_ExtractsRouteParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/reservations/:reservation_id/items/:sku", func(ctx *gin.Context) {
		fields := buildRequestLogFields(ctx)
		if fields["param_reservation_id"] != "res-1" {
			t.Fatalf("expected param_reservation_id, got %+v", fields)
		}
		if fields["reservation_id"] != "res-1" {
			t.Fatalf("expected reservation_id, got %+v", fields)
		}
		if fields["sku"] != "SKU-1" {
			t.Fatalf("expected sku, got %+v", fields)
		}
		ctx.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/reservations/res-1/items/SKU-1?foo=bar", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", w.Code)
	}
}
