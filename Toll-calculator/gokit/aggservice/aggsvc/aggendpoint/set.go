package aggendpoint

import (
	"context"

	"github.com/0x0Glitch/toll-calculator/gokit/aggservice/aggsvc/aggservice"
	"github.com/0x0Glitch/toll-calculator/types"
	"github.com/go-kit/kit/endpoint"
)

type Set struct {
	AggregateEndpoint endpoint.Endpoint
	CalculateEndpoint endpoint.Endpoint
}

type calculateRequest struct {
	OBUID int32 `json:"obuID"`
}

type aggregateRequest struct {
	Values float64 `json:"value"`
	OBUID  int32   `json:"obuID"`
	Unix   int64   `json:"unix"`
}

type CalculateResponse struct {
	OBUID         int32   `json:"obuID"`
	TotalDistance float64 `json:"totalDistance"`
	Amount        float64 `json:"amount"`
	Err           error   `json:"err"`
}

type AggregateResponse struct {
	Err error `json:"err"`
}

func (s Set) Aggregate(ctx context.Context, distance types.Distance) error {
	_, err := s.AggregateEndpoint(ctx, aggregateRequest{
		Values: distance.Values,
		OBUID:  distance.OBUID,
		Unix:   distance.Unix,
	})

	return err
}

func (s Set) Calculate(ctx context.Context, obuID int32) (*types.Invoice, error) {
	resp, err := s.CalculateEndpoint(ctx, calculateRequest{OBUID: obuID})
	if err != nil {
		return nil, err
	}
	result := resp.(CalculateResponse)

	return &types.Invoice{
		OBUID:         result.OBUID,
		TotalDistance: result.TotalDistance,
		Amount:        result.Amount,
	}, result.Err
}

func MakeAggregateEndpoint(s aggservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(aggregateRequest)
		err = s.Aggregate(ctx, types.Distance{
			Values: req.Values,
			OBUID:  req.OBUID,
			Unix:   req.Unix,
		})
		return AggregateResponse{Err: err}, nil
	}
}

// MakeConcatEndpoint constructs a Concat endpoint wrapping the service.
func MakeCalculateEndpoint(s aggservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(calculateRequest)
		v, err := s.Calculate(ctx, req.OBUID)
		return CalculateResponse{
			OBUID:         v.OBUID,
			TotalDistance: v.TotalDistance,
			Amount:        v.Amount,
			Err:           err,
		}, nil
	}
}
