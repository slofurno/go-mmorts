package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"runtime"
	"time"
)

func Tester() bool {
	return true
}

type Vector2 struct {
	X float64
	Y float64
}

func (v *Vector2) Normalize() *Vector2 {
	d := v.Length() //math.Sqrt(v.X*v.X + v.Y*v.Y)
	x := v.X / d
	y := v.Y / d
	return &Vector2{X: x, Y: y}

}

func (v *Vector2) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func NewVector2(x float64, y float64) *Vector2 {
	return &Vector2{X: x, Y: y}
}

func Distance(v1 *Vector2, v2 *Vector2) float64 {
	x := v2.X - v1.X
	y := v2.Y - v1.Y

	return math.Sqrt(x*x + y*y)
}

var connectionManager = NewWebsocketManager()

var connectedPlayers = make(map[string]*Websocket)
var commandQueue = make(chan Command, 500)
var squads = make(map[int]*Squad)
var ships = make(map[int]*Ship)

func processCommand(command Command) {
	fmt.Println("command type", reflect.TypeOf(command))
	switch command := command.(type) {
	case *MoveCommand:
		squad := squads[command.SquadId]
		squad.Position = &command.Target
	case *ReinforceCommand:
		squad := squads[command.SquadId]
		ship := ships[command.UnitId]

		if command.PlayerId != squad.Owner || command.PlayerId != ship.Owner {
			fmt.Println("wrong owner")
			return
		}
		//squad.Add(ship)
		squad.Add(ship)
		//ship.setSquad(squad)

	case *BuildCommand:

	default:
		fmt.Println("unrecoginized command type")
	}

}

func PrintShips() {
	//var buffer []byte
	buffer := []byte("{\"Ships\":[")

	index := 0

	for _, ship := range ships {
		if index > 0 {
			buffer = append(buffer, []byte(",")...)
		}
		s, _ := json.Marshal(ship)
		buffer = append(buffer, []byte(s)...)
		index++
	}
	buffer = append(buffer, []byte("],\"Squads\":[")...)

	index = 0
	for _, squad := range squads {
		if index > 0 {
			buffer = append(buffer, []byte(",")...)
		}
		s, _ := json.Marshal(squad)
		buffer = append(buffer, []byte(s)...)
		index++
	}

	buffer = append(buffer, []byte("]}")...)

	//fmt.Println(string(buffer))

	for ws := range connectionManager.Enumerate() {
		ws.Write([]byte(buffer))
	}

}

func Update() {
	for _, ship := range ships {
		if ship.squad != nil {
			heading := ship.GetHeading(ship.squad)
			ship.force.X = heading.X
			ship.force.Y = heading.Y
			//fmt.Println(heading)
			//ship.Position.X += heading.X
			//ship.Position.Y += heading.Y

		}

		for _, other := range ships {

			if ship != other {
				d := Distance(other.Position, ship.Position)
				if d <= 10 {
					v := other.GetHeading(ship)

					ship.force.X += v.X * ((10 - d) / 10)
					ship.force.Y += v.Y * ((10 - d) / 10)
				}
			}
		}

	}

	for _, ship := range ships {

		ship.Position.X += ship.force.X
		ship.Position.Y += ship.force.Y
	}

}

func Loop() {

	for {
		select {
		case command := <-commandQueue:
			processCommand(command)
		default:
			start := time.Now()
			Update()
			elapsed := time.Since(start)
			fmt.Printf("update elapsed: %s\r\n", elapsed)
			return
		}
	}

}

var _nextid int = 0

func nextId() int {
	_nextid++
	return _nextid
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	//testsquad := &Squad{Position: pos, Id: nextId()}

	squads[1] = NewSquad(200, 100, "asdf")
	squads[2] = NewSquad(300, 100, "asdf")
	squads[3] = NewSquad(400, 100, "asdf")
	squads[4] = NewSquad(500, 100, "asdf")
	squads[5] = NewSquad(100, 200, "asdf")
	squads[6] = NewSquad(100, 300, "asdf")
	squads[7] = NewSquad(100, 400, "asdf")
	squads[8] = NewSquad(100, 500, "asdf")
	squads[9] = NewSquad(600, 300, "asdf")
	squads[10] = NewSquad(100, 100, "asdf")

	for j := 1; j < 11; j++ {

		testships := []*Ship{}
		squad := squads[j]

		for i := 0; i < 50; i++ {

			n := NewShip(500, 500, "asdf")
			testships = append(testships, n)
			ships[n.Id] = n
		}
		squad.Add(testships...)
	}

	go func() {
		for {
			select {
			case <-time.After(time.Millisecond * 20):
				Loop()
				PrintShips()
			}

		}
	}()

	http.HandleFunc("/ws", WebsocketServer)
	http.ListenAndServe(":1616", nil)

}
