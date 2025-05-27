package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	"github.com/0x0Glitch/toll-calculator/types"
)

func main() {
	listenAddr := flag.String("listenaddr", ":3000", "the listen address of the HTTP server")
	store := NewMemoryStore()
	
		svc := NewInvoiceAggregator(store)
		svc = NewLogMiddleware(svc)
	
	makeHTTPTransport(*listenAddr, svc)
	fmt.Println("this is working fine")
}

func makeHTTPTransport(listenAddr string, svc Aggregator) {
	fmt.Println("HTTP transport running on port:", listenAddr)
	http.HandleFunc("/aggregate", handleAggregate(svc))
	http.ListenAndServe(listenAddr, nil)
}

func handleAggregate(svc Aggregator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var distance types.Distance
		if err := json.NewDecoder(r.Body).Decode(&distance); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if err := svc.AggregateDistance(distance); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}
}

func writeJSON(rw http.ResponseWriter, status int, v any) error {
	rw.WriteHeader(status)
	rw.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(rw).Encode(v)
}
