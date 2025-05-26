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
	var (
		svc = NewInvoiceAggregator(store)
	)
	makeHTTPTransport(*listenAddr,svc)
	fmt.Println("this is working fine")
}

func makeHTTPTransport(listenAddr string,svc Aggregator){
	fmt.Println("HTTP transport running on port:",listenAddr)
	http.HandleFunc("/aggregate", handleAggregate(svc))
	http.ListenAndServe(listenAddr,nil)
}

func handleAggregate(svc Aggregator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var distance types.Distance
		if err := json.NewDecoder(r.Body).Decode(&distance); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

}
