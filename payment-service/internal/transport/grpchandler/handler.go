package grpchandler

import (
	"context"

	pb "github.com/NoneNon9/convertedProto/payment/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"payment-service/internal/usecase"
)

type PaymentGRPCHandler struct {
	pb.UnimplementedPaymentServiceServer
	usecase *usecase.PaymentUseCase
}

func NewPaymentGRPCHandler(uc *usecase.PaymentUseCase) *PaymentGRPCHandler {
	return &PaymentGRPCHandler{usecase: uc}
}

func (h *PaymentGRPCHandler) ProcessPayment(ctx context.Context, req *pb.PaymentRequest) (*pb.PaymentResponse, error) {
	if req.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "Amount must be greater than 0")
	}

	payment, err := h.usecase.ProcessPayment(req.OrderId, req.Amount)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to process payment")
	}

	return &pb.PaymentResponse{
		Status:        payment.Status,
		TransactionId: payment.TransactionID,
	}, nil
}
