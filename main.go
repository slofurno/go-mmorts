package main

import (
	"fmt"
	"net/http"
	"time"
)

type Command struct {
	PlayerId string
	Unit     int
	Type     string
	Target   string
}

type Player struct {
	Out  chan string
	Done chan bool
}

type Vector2 struct {
	X float64
	Y float64
}

var connectedPlayers = make(map[string]*Player)
var commandQueue = make(chan *Command, 500)

func main() {

	go func() {
		for {
			select {
			case next := <-commandQueue:
				fmt.Println(next)
			case <-time.After(time.Second * 1):
				for _, value := range connectedPlayers {
					select {
					case value.Out <- "hi":
					default:
					}
				}
			}

		}
	}()

	http.HandleFunc("/ws", WebsocketServer)
	http.ListenAndServe(":1616", nil)

}
