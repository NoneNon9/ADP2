package usecase

import (
	"github.com/google/uuid"
	"payment-service/internal/domain"
)

type PaymentUseCase struct {
	repo domain.PaymentRepository
}

func NewPaymentUseCase(repo domain.PaymentRepository) *PaymentUseCase {
	return &PaymentUseCase{
		repo: repo,
	}
}

func (uc *PaymentUseCase) ProcessPayment(orderID string, amount int64) (domain.Payment, error) {
	status := "Authorized"
	if amount > 100000 {
		status = "Declined"
	}

	payment := domain.Payment{
		ID:            uuid.NewString(),
		OrderID:       orderID,
		TransactionID: "TXN-" + uuid.NewString(),
		Amount:        amount,
		Status:        status,
	}

	err := uc.repo.Save(payment)
	return payment, err
}

func (uc *PaymentUseCase) GetPaymentByOrderID(orderID string) (domain.Payment, error) {
	return uc.repo.GetByOrderID(orderID)
}
