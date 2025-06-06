package aggservice

import (
	"context"

	"github.com/0x0Glitch/toll-calculator/types"
)

type Service interface {
	Aggregate(context.Context, types.Distance) error
	Calculate(context.Context, int32) (*types.Invoice, error)
}


type BasicService struct {
	store Storer
}

func newBasicService(store Storer) Service{
	return &BasicService{
		store: store,
	}
}

func NewAggregatorService(store Storer) Service { 
	var svc Service
	svc = newBasicService(store)
	svc = newLoggingMiddleware()(svc)
	svc = newInstrumentationMiddleware()(svc)
	return svc
}

func (b *BasicService) Aggregate(_ context.Context, distance types.Distance) error {
	return b.store.Insert(&distance)
}

func (b *BasicService) Calculate(_ context.Context, obuID int32) (*types.Invoice, error) {
	distance, err := b.store.Get(obuID)
	if err != nil {
		return nil, err
	}
	return &types.Invoice{
		OBUID:         obuID,
		TotalDistance: distance,
		Amount:        distance * 9,
	}, nil
}