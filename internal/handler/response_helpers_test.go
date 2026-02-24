package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-inventory-reservations/internal/apperror"
	apimodel "go-inventory-reservations/internal/model/api"
	"go-inventory-reservations/internal/repository"

	"github.com/gin-gonic/gin"
)

func TestRequestIDFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	if got := requestIDFromContext(ctx); got != "" {
		t.Fatalf("expected empty request id, got %q", got)
	}

	ctx.Set(requestIDContextKey, "req-1")
	if got := requestIDFromContext(ctx); got != "req-1" {
		t.Fatalf("expected request id req-1, got %q", got)
	}
}

func TestWriteError_UsesRequestIDFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Set(requestIDContextKey, "req-123")

	writeError(ctx, http.StatusConflict, "Conflict", "CONFLICT", []apimodel.ErrorDetail{{Reason: "duplicate"}})

	if w.Code != http.StatusConflict {
		t.Fatalf("unexpected status: %d", w.Code)
	}

	var resp apimodel.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Success {
		t.Fatalf("expected unsuccessful response")
	}
	if resp.Error == nil || resp.Error.RequestID != "req-123" {
		t.Fatalf("expected request id in error object, got %+v", resp.Error)
	}
}

func TestWriteBindError_ValidationDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	type payload struct {
		SKU string `json:"sku" binding:"required"`
	}

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	var p payload
	err := ctx.ShouldBindJSON(&p)
	if err == nil {
		t.Fatalf("expected bind error")
	}

	writeBindError(ctx, err)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", w.Code)
	}
	var resp apimodel.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Error == nil || len(resp.Error.Details) == 0 {
		t.Fatalf("expected validation details, got %+v", resp.Error)
	}
}

func TestMapDomainError_AppError(t *testing.T) {
	err := apperror.New(apperror.CodeUnauthorized, apperror.CodeUnauthorizedMessage, "missing token")
	status, code, message, details := mapDomainError(err)

	if status != http.StatusUnauthorized || code != apperror.CodeUnauthorizedCode || message != apperror.CodeUnauthorizedMessage {
		t.Fatalf("unexpected mapping: status=%d code=%s message=%s", status, code, message)
	}
	if len(details) != 1 || details[0].Reason != "missing token" {
		t.Fatalf("unexpected details: %+v", details)
	}
}

func TestMapDomainError_ReservationVersionConflict(t *testing.T) {
	err := errors.Join(errors.New("wrapped"), repository.ErrReservationVersionConflict)
	status, code, message, details := mapDomainError(err)

	if status != http.StatusConflict {
		t.Fatalf("unexpected status: %d", status)
	}
	if code != apperror.CodeReservationVersionConflictCode {
		t.Fatalf("unexpected code: %s", code)
	}
	if message != apperror.CodeReservationVersionConflictMessage {
		t.Fatalf("unexpected message: %s", message)
	}
	if len(details) != 1 {
		t.Fatalf("expected one detail, got %+v", details)
	}
}

func TestMapDomainError_DBPatterns(t *testing.T) {
	status, code, message, _ := mapDomainError(sql.ErrNoRows)
	if status != http.StatusNotFound || code != apperror.CodeDBNoRowsCode || message != apperror.CodeDBNoRowsMessage {
		t.Fatalf("unexpected no rows mapping: %d %s %s", status, code, message)
	}

	status, code, message, _ = mapDomainError(errors.New("duplicate key value violates unique constraint"))
	if status != http.StatusConflict || code != apperror.CodeDBDuplicateKeyCode || message != apperror.CodeDBDuplicateKeyMessage {
		t.Fatalf("unexpected duplicate mapping: %d %s %s", status, code, message)
	}

	status, code, message, _ = mapDomainError(errors.New("violates check constraint"))
	if status != http.StatusUnprocessableEntity || code != apperror.CodeDBCheckConstraintCode || message != apperror.CodeDBCheckConstraintMessage {
		t.Fatalf("unexpected check mapping: %d %s %s", status, code, message)
	}
}

func TestMapDomainError_Defaults(t *testing.T) {
	status, code, message, details := mapDomainError(errors.New("some unexpected failure"))
	if status != http.StatusInternalServerError || code != apperror.CodeInternalErrorCode || message != apperror.CodeInternalErrorMessage {
		t.Fatalf("unexpected default mapping: %d %s %s", status, code, message)
	}
	if len(details) != 1 {
		t.Fatalf("expected one detail, got %+v", details)
	}
}
