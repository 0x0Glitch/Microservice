package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/0x0Glitch/toll-calculator/types"
	"github.com/gorilla/websocket"
)

const wsEndpoint = "ws://127.0.0.1:30000/ws"

var sendInterval = time.Second * 5

func genLatLong() (float64, float64) {
	return genCoord(), genCoord()
}

func genCoord() float64 {
	n := float64(rand.Intn(100) + 1)
	f := rand.Float64()
	return n + f
}

func main() {
	obuIDS := generateOBUIDS(20)
	conn, _, err := websocket.DefaultDialer.Dial(wsEndpoint, nil)
	if err != nil {
		log.Fatal(err)
	}

	for {
		for i := 0; i < len(obuIDS); i++ {
			lat, long := genLatLong()
			data := types.OBUData{
				OBUID: obuIDS[i],
				Lat:   lat,
				Long:  long,
			}
			if err := conn.WriteJSON(data); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%+v\n", data)
		}

		time.Sleep(sendInterval)
	}

}

func generateOBUIDS(n int) []int32 {
	ids := make([]int32, n)
	for i := 0; i < n; i++ {
		ids[i] = int32(rand.Intn(999999))
	}
	return ids
}

// func init() {
// 	if err := godotenv.Load(); err != nil {
// 		log.Fatal(err)
// 	}
// }
