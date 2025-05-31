package client

import (
	"context"

	"github.com/0x0Glitch/toll-calculator/types"
	"google.golang.org/grpc"
)

type GRPCClient struct {
	Endpoint string
	client   types.AggregatorClient
}

func NewGRPCClient(Endpoint string) *GRPCClient {

	conn, err := grpc.Dial(Endpoint, grpc.WithInsecure())
	if err != nil {
		return nil
	}
	c := types.NewAggregatorClient(conn)
	return &GRPCClient{
		Endpoint: Endpoint,
		client:   c,
	}
}

func (c *GRPCClient) Aggregate(ctx context.Context, req types.AggregatorRequest) error {
	_, err := c.client.Aggregate(ctx, &req)
	return err
}
