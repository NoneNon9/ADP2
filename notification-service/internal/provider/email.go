package provider

import (
	"errors"
	"log"
	"math/rand"
	"time"
)

type EmailSender interface {
	SendEmail(to string, orderID string, amount float64) error
}

type SimulatedEmailProvider struct{}

func NewSimulatedEmailProvider() *SimulatedEmailProvider {
	return &SimulatedEmailProvider{}
}

func (p *SimulatedEmailProvider) SendEmail(to string, orderID string, amount float64) error {

	time.Sleep(500 * time.Millisecond)

	if rand.Intn(100) < 30 {
		return errors.New("simulated transient network timeout")
	}

	log.Printf("[Notification Provider] Sent email to %s for Order #%s. Amount: $%.2f\n", to, orderID, amount)
	return nil
}
