package main

import (
	"math/rand"
)

type Planet struct {
	Id       int
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
	Id       int
	Position *Vector2
	velocity *Vector2
	force    *Vector2
	Owner    string
	SquadId  int
	squad    *Squad
}

func NewShip(x float64, y float64, owner string) *Ship {
	return &Ship{Position: &Vector2{X: x, Y: y}, Id: nextId(), force: &Vector2{}, Owner: owner}
}

func (s *Ship) GetPosition() *Vector2 {
	return s.Position
}

func (s *Ship) GetHeading(unit Unit) *Vector2 {
	dst := unit.GetPosition()

	dy := dst.Y - s.Position.Y
	dx := dst.X - s.Position.X

	if dy == 0 && dx == 0 {
		dy = rand.Float64()*2 - 1
		dx = rand.Float64()*2 - 1

	}

	delta := &Vector2{Y: dy, X: dx}
	return delta.Normalize()
}

func (ship *Ship) setSquad(squad *Squad) {

	if squad != nil {
		ship.squad = squad
		ship.SquadId = squad.Id
	} else {
		ship.squad = nil
		ship.SquadId = -1
	}
}

type Squad struct {
	Id       int
	Position *Vector2
	Owner    string
	//Ships    []*Ship
}

func NewSquad(x float64, y float64, owner string) *Squad {
	return &Squad{Position: &Vector2{X: x, Y: y}, Owner: owner, Id: nextId()}
}

func (squad *Squad) Add(ship ...*Ship) {
	for _, s := range ship {
		s.setSquad(squad)
	}
	//squad.Ships = append(squad.Ships, ship)
}

func (s *Squad) GetPosition() *Vector2 {
	return s.Position
}
