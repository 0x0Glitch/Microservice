package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/0x0Glitch/toll-calculator/types"
	"google.golang.org/grpc"
)

func main() {
	httplistenAddr := flag.String("httplistenaddr", ":3000", "the listen address of the HTTP server")
	grpclistenAddr := flag.String("grpclistenaddr", ":3001", "the listen address of the gRPC server")
	flag.Parse()

	store := NewMemoryStore()

	svc := NewInvoiceAggregator(store)
	svc = NewLogMiddleware(svc)

	// Start gRPC server in a separate goroutine
	serverErrCh := make(chan error, 1)
	go func() {
		fmt.Println("Starting gRPC server on", *grpclistenAddr)
		if err := makeGRPCTransport(*grpclistenAddr, svc); err != nil {
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

	// Try to connect to the gRPC server with retries
	// var c *client.GRPCClient
	// var err error
	// for i := 0; i < 5; i++ {
	// 	fmt.Printf("Attempting to connect to gRPC server (attempt %d)...\n", i+1)
	// 	c,err = client.NewGRPCClient(*grpclistenAddr)
	// 	if err == nil {
	// 		break
	// 	}
	// 	fmt.Printf("Connection attempt failed: %v. Retrying...\n", err)
	// 	time.Sleep(time.Second)
	// }

	// if err != nil {
	// 	log.Fatalf("Failed to connect to gRPC server after multiple attempts: %v", err)
	// }

	// fmt.Println("Successfully connected to gRPC server")

	// // Send test request
	// if _ = c.Aggregate(context.Background(), types.AggregatorRequest{
	// 	ObuID: 1,
	// 	Value: 58.55,
	// 	Unix:  time.Now().UnixNano(),
	// }); err != nil {
	// 	log.Fatalf("Failed to send aggregate request: %v", err)
	// }

	makeHTTPTransport(*httplistenAddr, svc)
	// fmt.Println("this is working fine")
}

func makeHTTPTransport(listenAddr string, svc Aggregator) {
	fmt.Println("HTTP transport running on port:", listenAddr)
	http.HandleFunc("/aggregate", handleAggregate(svc))
	http.HandleFunc("/invoice", handleGetInvoice(svc))
	http.ListenAndServe(listenAddr, nil)
}

func handleGetInvoice(svc Aggregator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
