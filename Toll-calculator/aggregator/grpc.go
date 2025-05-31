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
		OBUID:  int(req.ObuID),
		Values: req.Value,
		Unix:   req.Unix,
	}
	err := s.svc.AggregateDistance(distance)
	return &types.Empty{}, err
}

// This method is not used by the gRPC service but might be used internally
// func (s *GRPCAggregatorServer) AggregateDistance(req types.AggregatorRequest) error {
// 	distance := types.Distance{
// 		OBUID:  int(req.ObuID),
// 		Values: req.Value,
// 		Unix:   req.Unix,
// 	}
// 	return s.svc.AggregateDistance(distance)
// }
