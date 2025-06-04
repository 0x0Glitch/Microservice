package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/0x0Glitch/toll-calculator/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

type HTTPMetricHandler struct {
	reqCounter prometheus.Counter
	reqLatency prometheus.Histogram
	errCounter prometheus.Counter
}

type HTTPFunc func(http.ResponseWriter, *http.Request) error

type APIError struct {
	code int
	Err  error
}

func makeHTTPHandlerFunc(fn HTTPFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			if apiErr, ok := err.(APIError); ok {
				writeJSON(w, apiErr.code, map[string]string{"error": apiErr.Error()})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}
}

// Error implemets the error interface
func (e APIError) Error() string {
	return e.Err.Error()
}

func NewHTTPMetricHandler(reqName string) *HTTPMetricHandler {
	return &HTTPMetricHandler{
		reqCounter: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: fmt.Sprintf("http_%s_%s", reqName, "request_counter"),
			Name:      "aggregator",
		}),
		reqLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: fmt.Sprintf("http_%s_%s", reqName, "request_latency"),
			Name:      "aggregator",
			Buckets:   []float64{0.1, 0.5, 1},
		}),
		errCounter: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: fmt.Sprintf("http_%s_%s", reqName, "error_counter"),
			Name:      "aggregator",
		}),
	}
}

func (h *HTTPMetricHandler) Instrument(next HTTPFunc) HTTPFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var err error
		defer func(start time.Time) {
			latency := time.Since(start).Seconds()
			logrus.WithFields(logrus.Fields{
				"latency": latency,
				"request": r.RequestURI,
				"err":     err,
			}).Info("HTTP request latency")
			h.reqLatency.Observe(latency)
			h.reqCounter.Inc()
			if err != nil {
				h.errCounter.Inc()
			}
		}(time.Now())
		err = next(w, r)
		return err
	}
}

func handleGetInvoice(svc Aggregator) HTTPFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return APIError{
				code: http.StatusMethodNotAllowed,
				Err:  fmt.Errorf("invalid HTTP method %v", r.Method),
			}
		}
		values, ok := r.URL.Query()["obu"]
		// obuID := r.URL.Query()["obu"][0]
		if !ok {
			return APIError{
				code: http.StatusBadRequest,
				Err:  fmt.Errorf("missing OBU ID"),
			}

		}

		obuID, err := strconv.Atoi(values[0])
		if err != nil {
			return APIError{
				code: http.StatusBadRequest,
				Err:  fmt.Errorf("invalid OBU ID %v", values[0]),
			}
		}

		invoice, err := svc.CalculateInvoice(int32(obuID))
		if err != nil {
			return APIError{
				code: http.StatusInternalServerError,
				Err:  fmt.Errorf("failed to calculate invoice for OBU ID %v", obuID),
			}

		}
		return writeJSON(w, http.StatusOK, invoice)
	}
}

func handleAggregate(svc Aggregator) HTTPFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodPost {
			return APIError{
				code: http.StatusMethodNotAllowed,
				Err:  fmt.Errorf("invalid HTTP method %v", r.Method),
			}
		}
		var distance types.Distance
		if err := json.NewDecoder(r.Body).Decode(&distance); err != nil {
			return APIError{
				code: http.StatusBadRequest,
				Err:  fmt.Errorf("failed to decode distance: %v", err),
			}
		}
		if err := svc.AggregateDistance(&distance); err != nil {
			return APIError{
				code: http.StatusInternalServerError,
				Err:  fmt.Errorf("failed to aggregate distance: %v", err),
			}
		}
		return writeJSON(w, http.StatusOK, map[string]string{"message": "distance aggregated successfully"})
	}
}
