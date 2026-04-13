package grpchandler

import (
	pb "github.com/NoneNon9/convertedProto/order/v1"
	"order-service/internal/usecase"
)

type OrderGRPCHandler struct {
	pb.UnimplementedOrderTrackingServiceServer
	usecase *usecase.OrderUseCase
}

func NewOrderGRPCHandler(uc *usecase.OrderUseCase) *OrderGRPCHandler {
	return &OrderGRPCHandler{usecase: uc}
}

func (h *OrderGRPCHandler) SubscribeToOrderUpdates(req *pb.OrderRequest, stream pb.OrderTrackingService_SubscribeToOrderUpdatesServer) error {
	updateChan := h.usecase.Subscribe(req.OrderId)

	for status := range updateChan {
		err := stream.Send(&pb.OrderStatusUpdate{
			OrderId: req.OrderId,
			Status:  status,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
