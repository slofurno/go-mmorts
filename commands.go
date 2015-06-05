package main

import ()

type CommandType string

const (
	Move      CommandType = "MOV"
	Build     CommandType = "BLD"
	Reinforce CommandType = "ADD"
)

type Command interface {
	GetType() CommandType
}

type BuildCommand struct {
	PlanetId int
	PlayerId string
	UnitId   int
}

func (c BuildCommand) GetType() CommandType {
	return Build
}

type MoveCommand struct {
	SquadId  int
	PlayerId string
	Target   Vector2
}

func (c MoveCommand) GetType() CommandType {
	return Move
}

type ReinforceCommand struct {
	SquadId  int
	PlayerId string
	UnitId   int
}

func (c ReinforceCommand) GetType() CommandType {
	return Reinforce
}
