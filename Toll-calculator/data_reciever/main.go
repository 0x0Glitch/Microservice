package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/0x0Glitch/toll-calculator/types"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1028,
	WriteBufferSize: 1028,
}


type DataReceiver struct {
	msg  chan types.OBUData
	conn *websocket.Conn
	prod *kafka.Producer
}


var kafkaTopic = "obudata"



func main() {
	recv, err := NewDataReciever()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/ws", recv.wsHandler)
	http.ListenAndServe(":30000", nil)
}



func NewDataReciever() (*DataReceiver, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
	if err != nil {
		panic(err)
	}
	// start another go routine to check if we have delivered the data
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()
	return &DataReceiver{
		msg:  make(chan types.OBUData, 128),
		prod: p,
	}, nil
}



func (dr *DataReceiver) produceData(data types.OBUData) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = dr.prod.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &kafkaTopic,
			Partition: kafka.PartitionAny},
		Value: b,
	}, nil)
	return err
}



func (dr *DataReceiver) wsHandler(w http.ResponseWriter,r *http.Request) {
	fmt.Println("New OBU connected!")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatalf("websocket upgrade: %v", err)
		return
	}
	dr.conn = conn
	go dr.wsReceiveLoop() 

}




func (dr *DataReceiver) wsReceiveLoop() {
	for {
		var data types.OBUData
		if err := dr.conn.ReadJSON(&data); err != nil {
			log.Println("read error:", err)
			continue
		}
		fmt.Printf("data is %+v\n",data)
		if err := dr.produceData(data); err != nil {
			fmt.Println("kafka producer err:", err)
		}
	}
}
