package notifier

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"go-inventory-reservations/internal/logger"
	"go-inventory-reservations/internal/model"

	"github.com/google/uuid"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestNotifyQuoteExpired_NilReservation(t *testing.T) {
	n := NewQuoteExpirationNotifier("", time.Second, logger.New("error", "text"))
	err := n.NotifyQuoteExpired(context.Background(), nil)
	if err == nil {
		t.Fatalf("expected nil reservation error")
	}
}

func TestNotifyQuoteExpired_EmptyURLIsNoop(t *testing.T) {
	n := NewQuoteExpirationNotifier("", time.Second, logger.New("error", "text"))
	err := n.NotifyQuoteExpired(context.Background(), &model.Reservation{
		ReservationId: uuid.New(),
		QuoteId:       "Q-1",
		Status:        "EXPIRED",
	})
	if err != nil {
		t.Fatalf("expected nil error for empty endpoint, got %v", err)
	}
}

func TestNotifyQuoteExpired_SuccessAndFailureStatus(t *testing.T) {
	n := NewQuoteExpirationNotifier("http://notify.local/ok", time.Second, logger.New("error", "text"))
	impl := n.(*QuoteExpirationNotifier)
	impl.httpClient = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST, got %s", r.Method)
			}
			if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
				t.Fatalf("expected json content type, got %q", r.Header.Get("Content-Type"))
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	err := n.NotifyQuoteExpired(context.Background(), &model.Reservation{
		ReservationId: uuid.New(),
		QuoteId:       "Q-1",
		Status:        "EXPIRED",
	})
	if err != nil {
		t.Fatalf("unexpected success-path error: %v", err)
	}

	nFail := NewQuoteExpirationNotifier("http://notify.local/fail", time.Second, logger.New("error", "text"))
	failImpl := nFail.(*QuoteExpirationNotifier)
	failImpl.httpClient = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadGateway,
				Body:       io.NopCloser(strings.NewReader("bad gateway")),
				Header:     make(http.Header),
			}, nil
		}),
	}

	err = nFail.NotifyQuoteExpired(context.Background(), &model.Reservation{
		ReservationId: uuid.New(),
		QuoteId:       "Q-2",
		Status:        "EXPIRED",
	})
	if err == nil {
		t.Fatalf("expected error for non-2xx response")
	}
}
