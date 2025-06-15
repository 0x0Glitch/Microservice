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

func NewGRPCClient(Endpoint string) (*GRPCClient, error) {

	conn, err := grpc.Dial(Endpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	c := types.NewAggregatorClient(conn)
	return &GRPCClient{
		Endpoint: Endpoint,
		client:   c,
	}, nil
}

func (c *GRPCClient) Aggregate(ctx context.Context, req *types.AggregatorRequest) error {
	_, err := c.client.Aggregate(ctx, req)
	return err
}
