package domain

import "errors"

// PaymentEvent represents the incoming payload
type PaymentEvent struct {
	EventID       string  `json:"event_id"`
	OrderID       string  `json:"order_id"`
	Amount        float64 `json:"amount"`
	CustomerEmail string  `json:"customer_email"`
	Status        string  `json:"status"`
}

// ErrPermanentFailure is used to trigger our DLQ logic
var ErrPermanentFailure = errors.New("simulated permanent failure")

// NotificationStore defines the interface for idempotency checks
type NotificationStore interface {
	HasProcessed(eventID string) bool
	MarkProcessed(eventID string)
}
