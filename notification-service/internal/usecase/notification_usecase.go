package usecase

import (
	"context"
	"log"
	"time"

	"notification-service/internal/domain"
	"notification-service/internal/provider"
	"notification-service/internal/repository"
)

type NotificationUseCase struct {
	store    *repository.RedisIdempotencyStore
	provider provider.EmailSender
}

func NewNotificationUseCase(store *repository.RedisIdempotencyStore, provider provider.EmailSender) *NotificationUseCase {
	return &NotificationUseCase{store: store, provider: provider}
}

func (uc *NotificationUseCase) ProcessNotification(event domain.PaymentEvent) error {
	ctx := context.Background()

	if !uc.store.MarkAsProcessing(ctx, event.EventID) {
		log.Printf("Duplicate event ignored (Idempotent): %s", event.EventID)
		return nil
	}

	maxRetries := 3
	baseDelay := 2 * time.Second
	var err error

	for i := 0; i <= maxRetries; i++ {
		err = uc.provider.SendEmail(event.CustomerEmail, event.OrderID, event.Amount)
		if err == nil {
			return nil
		}

		if i < maxRetries {
			log.Printf("Provider failed: %v. Retrying in %v...", err, baseDelay)
			time.Sleep(baseDelay)
			baseDelay *= 2
		}
	}

	log.Printf("Message permanently failed after %d retries. Moving to DLQ.", maxRetries)
	return err
}
