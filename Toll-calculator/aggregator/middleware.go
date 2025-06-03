package main

import (
	"time"

	"github.com/0x0Glitch/toll-calculator/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

type LogMiddleware struct {
	next Aggregator
}

type MetricsMiddleware struct {
	next          Aggregator
	reqCounterAgg prometheus.Counter
	reqCounterInv prometheus.Counter
	reqLatencyAgg prometheus.Histogram
	reqLatencyInv prometheus.Histogram
	errCounterAgg prometheus.Counter
	errCounterInv prometheus.Counter
}

func NewMetricsMiddleware(next Aggregator) Aggregator {
	errCounterAgg := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "aggregator_error_counter",
		Name:      "aggregator",
	})
	errCounterInv := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "invoice_err_counter",
		Name:      "invoice",
	})

	reqCounterAgg := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "aggregator_request_counter",
		Name:      "aggregator",
	})
	reqCounterInv := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "invoice_request_counter",
		Name:      "invoice",
	})
	reqLatencyAgg := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "aggregator_request_latency",
		Name:      "aggregator_request_latency",
		Buckets:   []float64{0.1, 0.5, 1},
	})
	reqLatencyInv := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "invoice_request_latency",
		Name:      "invoice_request_latency",
		Buckets:   []float64{0.1, 0.5, 1},
	})
	return &MetricsMiddleware{
		errCounterAgg: errCounterAgg,
		errCounterInv: errCounterInv,
		next:          next,
		reqCounterAgg: reqCounterAgg,
		reqCounterInv: reqCounterInv,
		reqLatencyAgg: reqLatencyAgg,
		reqLatencyInv: reqLatencyInv,
	}
}

func (m *MetricsMiddleware) AggregateDistance(distance *types.Distance) (err error) {
	defer func(start time.Time) {
		m.reqCounterAgg.Inc()
		m.reqLatencyAgg.Observe(time.Since(start).Seconds())
		if err != nil {
			m.errCounterAgg.Inc()
		}
	}(time.Now())

	err = m.next.AggregateDistance(distance)
	return
}

func (m *MetricsMiddleware) CalculateInvoice(obuID int32) (inv *types.Invoice, err error) {
	defer func(start time.Time) {
		m.reqCounterInv.Inc()
		m.reqLatencyInv.Observe(time.Since(start).Seconds())
		if err != nil {
			m.errCounterInv.Inc()
		}
	}(time.Now())
	inv, err = m.next.CalculateInvoice(obuID)
	return
}

func NewLogMiddleware(next Aggregator) Aggregator {
	return &LogMiddleware{
		next: next,
	}
}
func (m *LogMiddleware) AggregateDistance(distance *types.Distance) (err error) {
	defer func(start time.Time) {
		logrus.WithFields(logrus.Fields{
			"time": time.Since(start),
			"err":  err,
		}).Info("Aggregate distance")
	}(time.Now())
	err = m.next.AggregateDistance(distance)
	return
}

func (m *LogMiddleware) CalculateInvoice(obuID int32) (inv *types.Invoice, err error) {
	defer func(start time.Time) {

		var (
			distance float64
			amount   float64
		)
		if inv != nil {
			distance = inv.TotalDistance
			amount = inv.Amount
		}
		logrus.WithFields(logrus.Fields{
			"time":     time.Since(start),
			"err":      err,
			"obuID":    obuID,
			"amount":   amount,
			"distance": distance,
		}).Info("Calculate Invoice")
	}(time.Now())
	inv, err = m.next.CalculateInvoice(obuID)
	return
}
