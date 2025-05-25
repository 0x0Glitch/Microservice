package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/0x0Glitch/toll-calculator/types"
	"github.com/gorilla/websocket"
)
var kafkaTopic = "obudata"
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1028,
	WriteBufferSize: 1028,
}

type DataReceiver struct {
	msg  chan types.OBUData
	conn *websocket.Conn
	prod DataProducer
}


func main() {
	recv, err := NewDataReciever()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/ws", recv.WsHandler)
	http.ListenAndServe(":30000", nil)
}

func (dr *DataReceiver) produceData(data types.OBUData) error {
	return dr.prod.ProduceData(data)
}

func NewDataReciever() (*DataReceiver, error) {
	var (
		p DataProducer
		err error
	)

	p, err = NewKafkaProducer(kafkaTopic)
	if err != nil {
		return nil, err
	}
	p = NewLogMiddleware(p)
	return &DataReceiver{
		msg:  make(chan types.OBUData, 128),
		prod: p,
	}, nil
}

func (dr *DataReceiver) WsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("New OBU connected!")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatalf("websocket upgrade: %v", err)
		return
	}
	dr.conn = conn
	go dr.WsReceiveLoop()

}

func (dr *DataReceiver) WsReceiveLoop() {
	for {
		var data types.OBUData
		if err := dr.conn.ReadJSON(&data); err != nil {
			log.Println("read error:", err)
			continue
		}
		fmt.Printf("data is %+v\n", data)
		if err := dr.prod.ProduceData(data); err != nil {
			fmt.Println("kafka producer err:", err)
		}
	}
}