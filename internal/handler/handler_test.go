package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go-inventory-reservations/internal/handler"
)

func TestHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &handler.Handlers{}

	tests := []struct {
		name           string
		method         string
		url            string
		handlerFunc    gin.HandlerFunc
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "CreateReservation",
			method:         "POST",
			url:            "/reservations",
			handlerFunc:    h.CreateReservation,
			expectedStatus: http.StatusOK,
			expectedMsg:    "reservation created",
		},
		{
			name:           "UpdateReservation",
			method:         "PUT",
			url:            "/reservations/1",
			handlerFunc:    h.UpdateReservation,
			expectedStatus: http.StatusOK,
			expectedMsg:    "reservation updated",
		},
		{
			name:           "GetReservation",
			method:         "GET",
			url:            "/reservations/1",
			handlerFunc:    h.GetReservation,
			expectedStatus: http.StatusOK,
			expectedMsg:    "get reservation success",
		},
		{
			name:           "GetReservationByQuoteId",
			method:         "GET",
			url:            "/reservations/quote/1",
			handlerFunc:    h.GetReservationByQuoteId,
			expectedStatus: http.StatusOK,
			expectedMsg:    "get reservation by quote ID success",
		},
		{
			name:           "GetReservationByOrderId",
			method:         "GET",
			url:            "/reservations/order/1",
			handlerFunc:    h.GetReservationByOrderId,
			expectedStatus: http.StatusOK,
			expectedMsg:    "get reservation by order ID success",
		},
		{
			name:           "DeleteReservation",
			method:         "DELETE",
			url:            "/reservations/1",
			handlerFunc:    h.DeleteReservation,
			expectedStatus: http.StatusOK,
			expectedMsg:    "reservation deleted",
		},
		{
			name:           "CommitReservation",
			method:         "POST",
			url:            "/reservations/1/commit",
			handlerFunc:    h.CommitReservation,
			expectedStatus: http.StatusOK,
			expectedMsg:    "reservation committed success",
		},
		{
			name:           "Attach",
			method:         "POST",
			url:            "/reservations/1/attach",
			handlerFunc:    h.Attach,
			expectedStatus: http.StatusOK,
			expectedMsg:    "reservation attached success",
		},
		{
			name:           "GetReservationAvailability",
			method:         "GET",
			url:            "/reservations/availability",
			handlerFunc:    h.GetReservationAvailability,
			expectedStatus: http.StatusOK,
			expectedMsg:    "reservation availability success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new router for each test
			router := gin.New()

			// Register just the one handler we're testing
			router.Handle(tt.method, tt.url, tt.handlerFunc)

			// Create a test request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.url, nil)

			// Serve the request
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check response body
			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			assert.Equal(t, tt.expectedMsg, response["message"])
		})
	}
}
