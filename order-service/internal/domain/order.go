package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidAmount    = errors.New("amount must be greater than 0")
	ErrCannotCancelPaid = errors.New("paid orders cannot be cancelled")
	ErrPaymentFailed    = errors.New("payment service unavailable or failed")
)

type Order struct {
	ID             string
	CustomerID     string
	ItemName       string
	Amount         int64
	Status         string // "Pending", "Paid", "Failed", "Cancelled"
	IdempotencyKey string
	CreatedAt      time.Time
}

type OrderRepository interface {
	Save(order Order) error
	GetByID(id string) (Order, error)
	GetByIdempotencyKey(key string) (Order, error)
	UpdateStatus(id string, status string) error
}

type PaymentGateway interface {
	AuthorizePayment(orderID string, amount int64) (status string, transactionID string, err error)
}
