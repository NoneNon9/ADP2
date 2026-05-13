package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"

	pb "github.com/NoneNon9/convertedProto/order/v1"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"order-service/internal/cache" // <-- Added cache import
	"order-service/internal/gateway"
	"order-service/internal/repository"
	"order-service/internal/transport/grpchandler"
	"order-service/internal/transport/httphandler"
	"order-service/internal/usecase"
)

func main() {
	dbConnStr := os.Getenv("DATABASE_URL")
	paymentServiceGRPC := os.Getenv("PAYMENT_SERVICE_GRPC_URL")
	grpcPort := os.Getenv("GRPC_PORT")
	redisURL := os.Getenv("REDIS_URL") // <-- New Redis ENV

	if grpcPort == "" {
		grpcPort = "50052"
	}
	if dbConnStr == "" {
		dbConnStr = "postgres://user:pass@localhost:5432/orderdb?sslmode=disable"
	}
	if redisURL == "" {
		redisURL = "localhost:6379"
	}
	paymentServiceURL := os.Getenv("PAYMENT_SERVICE_URL")
	if paymentServiceURL == "" {
		paymentServiceURL = "http://localhost:8081"
	}

	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("Database is unreachable: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})
	defer redisClient.Close()
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Redis is unreachable: %v", err)
	}

	orderRepo := repository.NewPostgresOrderRepository(db)
	orderCache := cache.NewRedisOrderCache(redisURL)

	paymentGateway, err := gateway.NewGRPCPaymentGateway(paymentServiceGRPC)
	if err != nil {
		log.Fatalf("Failed to dial payment service: %v", err)
	}

	orderUseCase := usecase.NewOrderUseCase(orderRepo, paymentGateway, orderCache)

	orderHTTPHandler := httphandler.NewOrderHandler(orderUseCase)
	router := httphandler.NewRouter(orderHTTPHandler)

	rateLimitedRouter := httphandler.RateLimiterMiddleware(redisClient, router)

	go func() {
		log.Println("Order Service HTTP starting on :8080...")
		if err := http.ListenAndServe(":8080", rateLimitedRouter); err != nil {
			log.Fatalf("HTTP Server failed: %v", err)
		}
	}()
	
	orderGRPCHandler := grpchandler.NewOrderGRPCHandler(orderUseCase)
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterOrderTrackingServiceServer(grpcServer, orderGRPCHandler)
	reflection.Register(grpcServer)
	log.Printf("Order Service gRPC starting on :%s...", grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC Server failed: %v", err)
	}
}
