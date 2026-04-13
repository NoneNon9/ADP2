package main

import (
	"context"
	"database/sql"
	pb "github.com/NoneNon9/convertedProto/payment/v1"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"payment-service/internal/transport/grpchandler"
	"time"

	_ "github.com/lib/pq"

	"payment-service/internal/repository"
	"payment-service/internal/usecase"
)

func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	log.Printf("Received request for %s", info.FullMethod)
	res, err := handler(ctx, req)
	log.Printf("Completed request for %s in %v. Error: %v", info.FullMethod, time.Since(start), err)
	return res, err
}

func main() {
	dbConnStr := os.Getenv("DATABASE_URL")
	if dbConnStr == "" {
		dbConnStr = "postgres://user:pass@localhost:5432/paymentdb?sslmode=disable"
	}
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}
	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Database is unreachable: %v", err)
	}

	paymentRepo := repository.NewPostgresPaymentRepository(db)
	paymentUseCase := usecase.NewPaymentUseCase(paymentRepo)
	paymentHandler := grpchandler.NewPaymentGRPCHandler(paymentUseCase)
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(loggingInterceptor))
	pb.RegisterPaymentServiceServer(grpcServer, paymentHandler)

	log.Printf("Payment Service gRPC starting on :%s...", grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
