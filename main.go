package main

import (
	"fmt"
	"math"
	"net/http"
	"time"
)

type Planet struct {
	Owner    string
	Position Vector2
}

type Player struct {
}

type Unit interface {
	GetPosition() *Vector2
}

type Movable interface {
	GetHeading(unit Unit)
}

type Ship struct {
	Position *Vector2
	Owner    string
	Squad    *Squad
}

func (s *Ship) GetHeading(unit Unit) *Vector2 {
	dst := unit.GetPosition()

	dy := dst.Y - s.Position.Y
	dx := dst.X - s.Position.X

	delta := &Vector2{Y: dy, X: dx}
	return delta.Normalize()
}

type Squad struct {
	Position *Vector2
	Owner    string
	Ships    []*Ship
}

func (squad *Squad) Add(ship *Ship) {
	squad.Ships = append(squad.Ships, ship)
}

func (s *Squad) GetPosition() *Vector2 {
	return s.Position
}

type Vector2 struct {
	X float64
	Y float64
}

func (v *Vector2) Normalize() *Vector2 {

	d := math.Sqrt(v.X*v.X + v.Y*v.Y)
	x := v.X / d
	y := v.Y / d
	return &Vector2{X: x, Y: y}

}

var connectedPlayers = make(map[string]*Websocket)
var commandQueue = make(chan Command, 500)
var squads = make(map[int]*Squad)
var ships = make(map[int]*Ship)

func processCommand(command Command) {
	switch command := command.(type) {
	case MoveCommand:
		squad := squads[command.SquadId]
		squad.Position = &command.Target
	case ReinforceCommand:
		squad := squads[command.SquadId]
		ship := ships[command.UnitId]

		if command.PlayerId != squad.Owner || command.PlayerId != ship.Owner {
			fmt.Println("wrong owner")
			return
		}
		//squad.Add(ship)
		ship.Squad = squad

	case BuildCommand:

	default:
		fmt.Println("unrecoginized command type")
	}

}

func PrintShips() {
	for _, ship := range ships {
		fmt.Println(ship.Position)
	}
}

func Update() {
	for _, ship := range ships {
		if ship.Squad != nil {
			heading := ship.GetHeading(ship.Squad)
			//fmt.Println(heading)
			ship.Position.X += heading.X
			ship.Position.Y += heading.Y

		}
	}

}

func Loop() {

	for {
		select {
		case command := <-commandQueue:
			processCommand(command)
		default:
			fmt.Println("updating")
			Update()
			return
		}
	}

}

func main() {

	pos := &Vector2{X: 1000, Y: 1000}
	pos1 := &Vector2{X: 0, Y: 0}
	pos2 := &Vector2{X: 2000, Y: 0}
	testsquad := &Squad{Position: pos}
	squads[1] = testsquad

	testship1 := &Ship{Squad: testsquad, Position: pos1}
	testship2 := &Ship{Squad: testsquad, Position: pos2}

	ships[1] = testship1
	ships[2] = testship2

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
