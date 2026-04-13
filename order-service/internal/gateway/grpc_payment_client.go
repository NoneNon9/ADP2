package gateway

import (
	"context"
	"time"

	pb "github.com/NoneNon9/convertedProto/payment/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCPaymentGateway struct {
	client pb.PaymentServiceClient
}

func NewGRPCPaymentGateway(targetURL string) (*GRPCPaymentGateway, error) {
	conn, err := grpc.Dial(targetURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &GRPCPaymentGateway{
		client: pb.NewPaymentServiceClient(conn),
	}, nil
}

func (g *GRPCPaymentGateway) AuthorizePayment(orderID string, amount int64) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := g.client.ProcessPayment(ctx, &pb.PaymentRequest{
		OrderId: orderID,
		Amount:  amount,
	})
	if err != nil {
		return "", "", err
	}
	return resp.Status, resp.TransactionId, nil
}
