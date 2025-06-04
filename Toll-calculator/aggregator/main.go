package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/0x0Glitch/toll-calculator/types"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	store := NewMemoryStore()
	svc := NewInvoiceAggregator(store)
	grpcListenAddr := os.Getenv("AGG_GRPC_LISTEN_ADDR")
	httpListenAddr := os.Getenv("AGG_HTTP_LISTEN_ADDR")

	svc = NewMetricsMiddleware(svc)
	svc = NewLogMiddleware(svc)

	// Start gRPC server in a separate goroutine
	serverErrCh := make(chan error, 1)
	go func() {
		fmt.Println("Starting gRPC server on", grpcListenAddr)
		if err := makeGRPCTransport(grpcListenAddr, svc); err != nil {
			serverErrCh <- err
		}
	}()

	// Give the server time to start
	fmt.Println("Waiting for gRPC server to start...")
	time.Sleep(time.Second * 2)

	// Check if server failed to start
	select {
	case err := <-serverErrCh:
		log.Fatalf("Failed to start gRPC server: %v", err)
	default:
		// Server started successfully or is still starting
	}
	makeHTTPTransport(httpListenAddr, svc)

}

func makeHTTPTransport(listenAddr string, svc Aggregator) {
	fmt.Println("HTTP transport running on port:", listenAddr)
	http.HandleFunc("/aggregate", handleAggregate(svc))
	http.HandleFunc("/invoice", handleGetInvoice(svc))
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(listenAddr, nil)
}

func handleGetInvoice(svc Aggregator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		values, ok := r.URL.Query()["obu"]
		// obuID := r.URL.Query()["obu"][0]
		if !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing OBU ID"})
			return
		}

		obuID, err := strconv.Atoi(values[0])
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid OBU ID"})
			return
		}

		invoice, err := svc.CalculateInvoice(int32(obuID))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, invoice)
	}
}

func makeGRPCTransport(listenAddr string, svc Aggregator) error {
	// make a TCP listener
	ln, err := net.Listen("tcp", listenAddr)

	if err != nil {
		return err
	}
	defer ln.Close()
	// Make a new GRPC native server with options
	server := grpc.NewServer([]grpc.ServerOption{}...)
	//Register our GRPC server implementation to the GRPC package
	types.RegisterAggregatorServer(server, NewAggregatorGRPCServer(svc))
	return server.Serve(ln)
}

func handleAggregate(svc Aggregator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var distance types.Distance
		if err := json.NewDecoder(r.Body).Decode(&distance); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if err := svc.AggregateDistance(&distance); err != nil {
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


	

