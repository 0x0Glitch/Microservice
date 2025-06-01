package main

import (
	"time"

	"github.com/0x0Glitch/toll-calculator/types"
	"github.com/sirupsen/logrus"
)

type LogMiddleware struct {
	next Aggregator
}

func NewLogMiddleware(next Aggregator) Aggregator {
	return &LogMiddleware{
		next: next,
	}
}
func (m *LogMiddleware) AggregateDistance(distance types.Distance) (err error) {
	defer func(start time.Time) {
		logrus.WithFields(logrus.Fields{
			"time": time.Since(start),
			"err":  err,
		}).Info("Aggregate distance")
	}(time.Now())
	err = m.next.AggregateDistance(distance)
	return
}

func (m *LogMiddleware) CalculateInvoice(obuID uint64) (inv *types.Invoice, err error) {
	defer func(start time.Time) {

		var (
			distance float64
			amount   float64
		)
		if inv != nil{
			distance = inv.TotalDistance
			amount = inv.Amount
		}
		logrus.WithFields(logrus.Fields{
			"time": time.Since(start),
			"err":  err,
			"obuID":obuID,
			"amount":amount,
			"distance":distance,
		}).Info("Calculate Invoice")
	}(time.Now())
	inv,err = m.next.CalculateInvoice(obuID)
	return
}
 