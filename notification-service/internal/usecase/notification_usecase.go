package usecase

import (
	"log"
	"notification-service/internal/domain"
)

type NotificationUseCase struct {
	store domain.NotificationStore
}

func NewNotificationUseCase(store domain.NotificationStore) *NotificationUseCase {
	return &NotificationUseCase{store: store}
}

func (uc *NotificationUseCase) ProcessNotification(event domain.PaymentEvent) error {
	if event.OrderID == "FAIL_ME" {
		log.Printf("Simulating permanent failure for Order %s. Moving to DLQ...", event.OrderID)
		return domain.ErrPermanentFailure
	}

	if uc.store.HasProcessed(event.EventID) {
		log.Printf("Duplicate event ignored: %s", event.EventID)
		return nil
	}
	
	log.Printf("[Notification] Sent email to %s for Order #%s. Amount: $%.2f\n",
		event.CustomerEmail, event.OrderID, event.Amount)

	uc.store.MarkProcessed(event.EventID)
	return nil
}
