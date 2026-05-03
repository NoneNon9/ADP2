package broker

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"notification-service/internal/domain"
	"notification-service/internal/usecase"
)

type RabbitMQConsumer struct {
	conn    *amqp.Connection
	ch      *amqp.Channel
	usecase *usecase.NotificationUseCase
}

func NewRabbitMQConsumer(url string, uc *usecase.NotificationUseCase) (*RabbitMQConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	_ = ch.ExchangeDeclare("dlx_exchange", "direct", true, false, false, false, nil)
	_, _ = ch.QueueDeclare("payment_dlq", true, false, false, false, nil)
	_ = ch.QueueBind("payment_dlq", "dlx_routing_key", "dlx_exchange", false, nil)

	_ = ch.ExchangeDeclare("payment_exchange", "topic", true, false, false, false, nil)
	args := amqp.Table{
		"x-dead-letter-exchange":    "dlx_exchange",
		"x-dead-letter-routing-key": "dlx_routing_key",
	}
	q, _ := ch.QueueDeclare("payment.completed.queue", true, false, false, false, args)
	_ = ch.QueueBind(q.Name, "payment.completed", "payment_exchange", false, nil)

	return &RabbitMQConsumer{conn: conn, ch: ch, usecase: uc}, nil
}

func (c *RabbitMQConsumer) Start() error {
	msgs, err := c.ch.Consume(
		"payment.completed.queue",
		"notification-service",
		false,
		false, false, false, nil,
	)
	if err != nil {
		return err
	}

	log.Printf(" [*] Notification Service waiting for messages.")

	go func() {
		for d := range msgs {
			c.processMessage(d)
		}
	}()
	return nil
}

func (c *RabbitMQConsumer) processMessage(d amqp.Delivery) {
	var event domain.PaymentEvent
	if err := json.Unmarshal(d.Body, &event); err != nil {
		log.Printf("Error unmarshalling message: %v", err)
		d.Nack(false, false)
		return
	}

	err := c.usecase.ProcessNotification(event)

	if err == domain.ErrPermanentFailure {
		d.Nack(false, false)
		return
	}

	d.Ack(false)
}

func (c *RabbitMQConsumer) Close() {
	c.ch.Close()
	c.conn.Close()
}
