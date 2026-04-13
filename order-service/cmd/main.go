package main

import (
	"database/sql"
	pb "github.com/NoneNon9/convertedProto/order/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	"order-service/internal/transport/grpchandler"
	"os"

	_ "github.com/lib/pq"

	"order-service/internal/gateway"
	"order-service/internal/repository"
	"order-service/internal/transport/httphandler"
	"order-service/internal/usecase"
)

func main() {
	dbConnStr := os.Getenv("DATABASE_URL")
	paymentServiceGRPC := os.Getenv("PAYMENT_SERVICE_GRPC_URL")
	grpcPort := os.Getenv("GRPC_PORT")

	if grpcPort == "" {
		grpcPort = "50052"
	}
	if dbConnStr == "" {
		dbConnStr = "postgres://user:pass@localhost:5432/orderdb?sslmode=disable"
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

	orderRepo := repository.NewPostgresOrderRepository(db)

	paymentGateway, err := gateway.NewGRPCPaymentGateway(paymentServiceGRPC)
	if err != nil {
		log.Fatalf("Failed to dial payment service: %v", err)
	}

	orderUseCase := usecase.NewOrderUseCase(orderRepo, paymentGateway)

	orderHTTPHandler := httphandler.NewOrderHandler(orderUseCase)
	router := httphandler.NewRouter(orderHTTPHandler)
	go func() {
		log.Println("Order Service HTTP starting on :8080...")
		if err := http.ListenAndServe(":8080", router); err != nil {
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
