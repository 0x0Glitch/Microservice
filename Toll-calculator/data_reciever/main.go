package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/0x0Glitch/toll-calculator/types"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1028,
	WriteBufferSize: 1028,
}

type DataReceiver struct {
	msg  chan types.OBUData
	conn *websocket.Conn
}

func NewDataReciever() *DataReceiver{
	return &DataReceiver{
		msg: make(chan types.OBUData,128),
	}
}

func main() {
	recv := NewDataReciever()
	http.HandleFunc("/ws",recv.wsHandler)
	fmt.Println("data reciever working fine")
	http.ListenAndServe(":30000",nil)
}


func (dr *DataReceiver) wsHandler(w http.ResponseWriter, r *http.Request) {
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
			log.Println(err)
			continue
		}
		fmt.Printf("Recieved OBU data from [%d]:: <lat %.2f,long %.2f>\n",data.OBUID,data.Lat,data.Long)
		dr.msg <- data
	}
}
