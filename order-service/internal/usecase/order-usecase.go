package usecase

import (
	"context"
	"github.com/google/uuid"
	"order-service/internal/cache"
	"order-service/internal/domain"
	"sync"
	"time"
)

type OrderUseCase struct {
	repo        domain.OrderRepository
	payment     domain.PaymentGateway
	cache       *cache.RedisOrderCache
	subscribers map[string]chan string
	mu          sync.RWMutex
}

func NewOrderUseCase(repo domain.OrderRepository, payment domain.PaymentGateway, cache *cache.RedisOrderCache) *OrderUseCase {
	return &OrderUseCase{
		repo:        repo,
		payment:     payment,
		cache:       cache,
		subscribers: make(map[string]chan string),
	}
}

func (uc *OrderUseCase) notify(orderID string, status string) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	if ch, exists := uc.subscribers[orderID]; exists {
		ch <- status
	}
}

func (uc *OrderUseCase) Subscribe(orderID string) <-chan string {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	if _, exists := uc.subscribers[orderID]; !exists {
		uc.subscribers[orderID] = make(chan string, 10)
	}
	return uc.subscribers[orderID]
}

func (uc *OrderUseCase) CreateOrder(customerID, itemName string, amount int64, idempotencyKey string) (domain.Order, error) {

	if amount <= 0 {
		return domain.Order{}, domain.ErrInvalidAmount
	}

	order := domain.Order{
		ID:             uuid.NewString(),
		CustomerID:     customerID,
		ItemName:       itemName,
		Amount:         amount,
		Status:         "Pending",
		IdempotencyKey: idempotencyKey,
		CreatedAt:      time.Now(),
	}
	if err := uc.repo.Save(order); err != nil {
		return domain.Order{}, err
	}

	uc.notify(order.ID, "Pending")

	status, _, err := uc.payment.AuthorizePayment(order.ID, order.Amount)
	if err != nil {
		uc.repo.UpdateStatus(order.ID, "Failed")
		uc.notify(order.ID, "Failed")
		return domain.Order{}, domain.ErrPaymentFailed
	}

	finalStatus := "Paid"
	if status == "Declined" {
		finalStatus = "Failed"
	}

	uc.repo.UpdateStatus(order.ID, finalStatus)
	order.Status = finalStatus
	uc.notify(order.ID, finalStatus)

	return order, nil
}

func (uc *OrderUseCase) CancelOrder(id string) error {
	order, err := uc.repo.GetByID(id)
	if err != nil {
		return err
	}
	if order.Status == "Paid" {
		return domain.ErrCannotCancelPaid
	}

	err = uc.repo.UpdateStatus(id, "Cancelled")
	if err == nil {
		uc.notify(id, "Cancelled")
	}
	return err
}

func (uc *OrderUseCase) GetOrder(ctx context.Context, id string) (domain.Order, error) {

	order, err := uc.cache.Get(ctx, id)
	if err == nil {
		return order, nil
	}

	order, err = uc.repo.GetByID(id)
	if err != nil {
		return order, err
	}

	_ = uc.cache.Set(ctx, order)
	return order, nil
}
