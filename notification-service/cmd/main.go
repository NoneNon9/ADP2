package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"notification-service/internal/broker"
	"notification-service/internal/provider"
	"notification-service/internal/repository"
	"notification-service/internal/usecase"
)

func main() {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	providerMode := os.Getenv("PROVIDER_MODE")
	if providerMode == "" {
		providerMode = "SIMULATED"
	}

	log.Printf("Connecting to Redis at %s...", redisURL)
	store := repository.NewRedisIdempotencyStore(redisURL)

	var emailProvider provider.EmailSender
	if providerMode == "SIMULATED" {
		log.Println("Initializing SIMULATED Email Provider...")
		emailProvider = provider.NewSimulatedEmailProvider()
	} else {
		log.Println("Initializing REAL Email Provider (Fallback to Simulated)...")
		emailProvider = provider.NewSimulatedEmailProvider()
	}

	notificationUseCase := usecase.NewNotificationUseCase(store, emailProvider)

	consumer, err := broker.NewRabbitMQConsumer(rabbitURL, notificationUseCase)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ consumer: %v", err)
	}
	defer consumer.Close()
	
	if err := consumer.Start(); err != nil {
		log.Fatalf("Failed to start consuming: %v", err)
	}

	log.Println("Notification Worker started successfully. Press CTRL+C to exit.")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Notification Service gracefully...")
}
