package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"notification-service/internal/broker"
	"notification-service/internal/repository"
	"notification-service/internal/usecase"
)

func main() {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	store := repository.NewInMemoryStore()
	notificationUseCase := usecase.NewNotificationUseCase(store)
	consumer, err := broker.NewRabbitMQConsumer(rabbitURL, notificationUseCase)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ consumer: %v", err)
	}
	defer consumer.Close()
	if err := consumer.Start(); err != nil {
		log.Fatalf("Failed to start consuming: %v", err)
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Notification Service gracefully...")
}
