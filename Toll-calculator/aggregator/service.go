package main

import (
	"fmt"

	"github.com/0x0Glitch/toll-calculator/types"
)

type Aggregator interface {
	AggregateDistance(*types.Distance) error
	CalculateInvoice(int32) (*types.Invoice, error)
}

type Storer interface {
	Insert(*types.Distance) error
	Get(int32) (float64, error)
}

type InvoiceAggregator struct {
	store Storer
}

const basePrice = 315


func NewInvoiceAggregator(store Storer) Aggregator {
	return &InvoiceAggregator{
		store: store,
	}
}
func (i *InvoiceAggregator) AggregateDistance(distance *types.Distance) error {
	fmt.Println("processing and inserting distance in the storage:", distance)
	return i.store.Insert(distance)
}

func (i *InvoiceAggregator) CalculateInvoice(obuID int32) (*types.Invoice, error) {
	dist, err := i.store.Get(obuID)
	if err != nil {
		return nil, err
	}
	inv := &types.Invoice{
		OBUID:         obuID,
		TotalDistance: dist,
		Amount:        basePrice * dist,
	}
	return inv, nil
}
