package main

import (
	"net"

	uuid "github.com/satori/go.uuid"
)

// PlayerState used below
type PlayerState int

// Constants
const (
	IDLE PlayerState = iota + 1
	DEAD
	MOVING
	JUMPING
	FIRING
)

// Vector is a vector
type Vector []float32

// Player is a player
type Player struct {
	connection net.Conn
	health     float32
	id         uuid.UUID
	name       string
	state      PlayerState
}

func newPlayer(name string, conn net.Conn) Player {
	return Player{
		connection: conn,
		health:     100.00,
		id:         uuid.Must(uuid.NewV4()),
		name:       name,
		state:      IDLE,
	}
}

// Hit for when a player hit events is reported
// TODO: can we determine if the players are close enough to actually hit each other?
func (p *Player) Hit(damage float32) bool {
	p.health -= damage
	p.state = DEAD
	return p.IsDead()
}

// IsDead to detect if player is dead
func (p *Player) IsDead() bool {
	if p.state == DEAD {
		return true
	}
	return false
}

// Move updates server's data for a player
// TODO: try to protect against teleporting, flying etc.
// RETURNS: bool indicating whether or not the client should be told the move was approved
func (p *Player) Move(v Vector) bool {
	return true
}

// ID return player UUID
func (p *Player) ID() uuid.UUID {
	return p.id
}

// Health return player health
func (p *Player) Health() float32 {
	return p.health
}

// Name return player name
func (p *Player) Name() string {
	return p.name
}

// Connection return player connection
func (p *Player) Connection() net.Conn {
	return p.connection
}

// SetName change player's name
func (p *Player) SetName(name string) {
	p.name = name
}

// TakeDamage player gets hurt
func (p *Player) TakeDamage(amount float32) {
	p.health -= amount
	if p.health < 0 {
		p.health = 0
	}
}

// Respawn player stats
func (p *Player) Respawn(health float32) {
	p.health = health
	p.state = IDLE
	prettyPrint(p)
}
