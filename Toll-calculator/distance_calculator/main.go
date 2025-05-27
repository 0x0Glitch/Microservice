package main

import (
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
	)
	svc = NewCalculatorService()
	svc = NewLogMiddleware(svc)
	
	KafkaConsumer, err := NewKafkaConsumer(kafkaTopic,svc,client.NewClient("http://127.0.0.1:3000/aggregate"))
	if err != nil {
		log.Fatal(err)
	}
	KafkaConsumer.Start()
	fmt.Println("everything working fine")
}
