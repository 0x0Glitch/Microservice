package main

import (
	"fmt"
	"log"
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
	
	KafkaConsumer, err := NewKafkaConsumer(kafkaTopic,svc)
	if err != nil {
		log.Fatal(err)
	}
	KafkaConsumer.Start()
	fmt.Println("everything working fine")
}
