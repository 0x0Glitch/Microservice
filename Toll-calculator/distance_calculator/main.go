package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/0x0Glitch/toll-calculator/aggregator/client"
)

//	type DistanceCalculator struct {
//		consumer DataConsumer
//	}
const kafkaTopic = "obudata"

func main() {
	var (
		svc CalculatorServicer
		aggregateEndpoint = "http://127.0.0.1:3000"
	)
	// httplistenAddr := flag.String("httplistenaddr", ":3001", "the listen address of the gRPC server")
	flag.Parse()
	svc = NewCalculatorService()
	svc = NewLogMiddleware(svc)
	c := client.NewHTTPClient(aggregateEndpoint)
	
	KafkaConsumer, err := NewKafkaConsumer(kafkaTopic, svc,c)
	if err != nil {
		log.Fatal(err)
	}
	KafkaConsumer.Start()
	fmt.Println("everything working fine")
}
