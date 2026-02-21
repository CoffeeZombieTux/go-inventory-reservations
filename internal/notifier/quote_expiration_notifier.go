package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go-inventory-reservations/internal/logger"
	"go-inventory-reservations/internal/model"
)

const maxResponseBodyLogBytes = 1024

// QuoteExpirationNotifierInterface notifies external services about quote expiration.
type QuoteExpirationNotifierInterface interface {
	NotifyQuoteExpired(ctx context.Context, reservation *model.Reservation) error
}

// QuoteExpirationNotifier sends quote expiration webhooks.
type QuoteExpirationNotifier struct {
	endpointURL string
	httpClient  *http.Client
	logger      *logger.Logger
}

type quoteExpirationPayload struct {
	QuoteID       string    `json:"quote_id"`
	ReservationID string    `json:"reservation_id"`
	Status        string    `json:"status"`
	ExpiredAt     time.Time `json:"expired_at"`
}

// NewQuoteExpirationNotifier creates a new notifier instance.
func NewQuoteExpirationNotifier(endpointURL string, timeout time.Duration, log *logger.Logger) QuoteExpirationNotifierInterface {
	return &QuoteExpirationNotifier{
		endpointURL: strings.TrimSpace(endpointURL),
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: log,
	}
}

// NotifyQuoteExpired sends expiration notification for a reservation quote.
func (n *QuoteExpirationNotifier) NotifyQuoteExpired(ctx context.Context, reservation *model.Reservation) error {
	if reservation == nil {
		return fmt.Errorf("reservation is nil")
	}
	if n.endpointURL == "" {
		return nil
	}

	payload := quoteExpirationPayload{
		QuoteID:       reservation.QuoteId,
		ReservationID: reservation.ReservationId.String(),
		Status:        reservation.Status,
		ExpiredAt:     time.Now().UTC(),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal quote expiration payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.endpointURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build quote expiration request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	n.logger.WithFields(logger.Fields{
		"request_url":    n.endpointURL,
		"quote_id":       payload.QuoteID,
		"reservation_id": payload.ReservationID,
		"status":         payload.Status,
		"request_body":   string(body),
	}).Info(logger.LogMessageQuoteExpirationNotifyRequest)

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send quote expiration request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			n.logger.WithError(closeErr).WithFields(logger.Fields{
				"request_url": n.endpointURL,
			}).Error(logger.LogMessageFailedToCloseResponseBody)
		}
	}()

	responseBodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return fmt.Errorf("read quote expiration response body: %w", readErr)
	}

	responseBody := strings.TrimSpace(string(responseBodyBytes))
	responseBodyTruncated := false
	if len(responseBody) > maxResponseBodyLogBytes {
		responseBody = responseBody[:maxResponseBodyLogBytes]
		responseBodyTruncated = true
	}

	n.logger.WithFields(logger.Fields{
		"request_url":             n.endpointURL,
		"quote_id":                payload.QuoteID,
		"reservation_id":          payload.ReservationID,
		"response_code":           resp.StatusCode,
		"response_body":           responseBody,
		"response_body_truncated": responseBodyTruncated,
	}).Info(logger.LogMessageQuoteExpirationNotifyResponse)

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("unexpected quote expiration response status: %d", resp.StatusCode)
	}

	return nil
}
