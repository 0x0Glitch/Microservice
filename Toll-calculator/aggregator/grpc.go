package main

import (
	"context"

	"github.com/0x0Glitch/toll-calculator/types"
)

type GRPCAggregatorServer struct {
	types.UnimplementedAggregatorServer
	svc Aggregator
}

func NewAggregatorGRPCServer(svc Aggregator) *GRPCAggregatorServer {
	return &GRPCAggregatorServer{
		svc: svc,
	}
}

// Aggregate implements the Aggregate RPC method from the protobuf definition
func (s *GRPCAggregatorServer) Aggregate(ctx context.Context, req *types.AggregatorRequest) (*types.Empty, error) {
	distance := types.Distance{
		OBUID:  int32(req.ObuID),
		Values: req.Value,
		Unix:   req.Unix,
	}
	err := s.svc.AggregateDistance(&distance)
	return &types.Empty{}, err
}

func (s *GRPCAggregatorServer) CalculateInvoice(ctx context.Context, req *types.Invoice) (*types.Invoice, error) {
	return s.svc.CalculateInvoice(int32(req.OBUID))
}
