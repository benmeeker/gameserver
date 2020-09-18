package main

import (
	"fmt"
	"log"
	"time"

	uuid "github.com/satori/go.uuid"
)

// ServerState used below
type ServerState int

// ServerState constants
const (
	STOPPED = iota + 1
	STARTED
)

// Server object
type Server struct {
	ID         uuid.UUID
	Started    time.Time
	Roster     Roster
	State      ServerState
	Duration   float64
	Timed      bool
	MaxKills   int
	MaxPlayers int
}

func newServer(timed bool, durationSeconds float64) Server {
	return Server{
		Roster:     newRoster(),             // Where the server keeps player states
		Timed:      timed,                   // TODO: use this along with Duration
		Duration:   durationSeconds,         // TODO: use this
		ID:         uuid.Must(uuid.NewV4()), // UUID for the server (send this to clients on connect)
		MaxPlayers: 20,
	}
}

// Start the server
func (s *Server) Start() {
	s.Started = time.Now()
	s.State = STARTED
	log.Print("Server started")
}

// End the server
// TODO: instruct clients to cease playability
// TODO: disconnect all player connections and wipe out roster??? or just exit?
func (s *Server) End() {
	log.Print("Server stopped")
	s.State = STOPPED
}

// TimeCheck to see if the server needs to exit
func (s *Server) TimeCheck() bool {
	var timeLeft = time.Since(s.Started).Seconds() - s.Duration
	if timeLeft <= 1 {
		log.Print("Server time is up!!!")
		return false
	}
	log.Printf("Server has %.2f seconds left", timeLeft)
	return true
}

// PlayerJoin for when a player joins the server
func (s *Server) PlayerJoin(p Player) {
	var m Message
	if ROSTER.PlayerCount() >= s.MaxPlayers {
		log.Printf("Player %s join failed, server full", p.Name())

		return
	}
	if ROSTER.AddPlayer(p) {
		pname := ROSTER.PlayerName(p.ID())
		msgText := fmt.Sprintf("Player %s joined", pname) // lookup the name as it may have changed in AddPlayer()
		log.Printf(msgText)
		s.BroadcastMessage(newBroadcastMessage(msgText))
		s.SendScores()
		m = newJoinMessage(p.ID().String(), pname, "success")
	} else {
		log.Printf("Player join failed (non-unique UUID or remoteAddr:port)")
		m = newJoinMessage("", "", "failure")
	}
	netSendMessage(p.Connection(), m)
}

// PlayerHit for when one player hits another
// TODO: send hit message to victim
func (s *Server) PlayerHit(victim string, aUUID uuid.UUID, damage float32) {
	log.Printf("processing player hit %s", victim)
	vUUID := ROSTER.PlayerLookup(victim, nil)
	log.Printf("processing player hit uuid %s", vUUID.String())

	if vUUID == uuid.Nil {
		log.Printf("%s cannot hit non-existent player %s", ROSTER.PlayerName(aUUID), victim)
		return
	}
	health := ROSTER.PlayerDamage(vUUID, damage)

	// notify player of remaining health
	m := newClientMessage(fmt.Sprintf("remaining health %3.2f", health))
	s.ClientMessage(vUUID, m)

	if health <= 0 {
		s.PlayerKilled(vUUID, aUUID)
	}
}

// PlayerKilled registers a player kill
// TODO: tell victim they're dead
func (s *Server) PlayerKilled(vUUID, aUUID uuid.UUID) {
	ROSTER.PlayerKilled(vUUID, aUUID)
	m := newKillMessage(ROSTER.PlayerName(aUUID))
	s.ClientMessage(vUUID, m)

	msgText := fmt.Sprintf("%s killed %s", ROSTER.PlayerName(aUUID), ROSTER.PlayerName(vUUID))
	log.Print(msgText)

	s.BroadcastMessage(newBroadcastMessage(msgText))
	s.SendScores()
	ROSTER.RespawnPlayer(vUUID)
}

// BroadcastMessage is used to send a message to all players
// TODO: remove players from roster if they error on a broadcast message?
func (s *Server) BroadcastMessage(m Message) {
	for _, conn := range ROSTER.PlayerConnAll() {
		err := netSendMessage(conn, m)
		if err != nil {
			log.Printf("BROADCAST error sending to %s %s", conn.RemoteAddr().String(), err.Error())
		}
	}
}

// SendScores broadcasts a message with the new scores to all players
// TODO: integrate with new roster
func (s *Server) SendScores() {
	var scores = ROSTER.Scoreboard()

	prettyPrint(scores)

	msg := newScoreboardMessage()
	s.BroadcastMessage(msg)
}

// Reset resets the server
func (s *Server) Reset() {
	s.End()
	ROSTER.Reset()
	s.Start()
}

// ClientMessage send a message to client (not chat)
func (s *Server) ClientMessage(puuid uuid.UUID, m Message) {
	conn := ROSTER.PlayerConn(puuid)
	//m := newClientMessage(content)
	err := netSendMessage(conn, m)
	if err != nil {
		log.Printf("CLIENT error sending to %s %s", conn.RemoteAddr().String(), err.Error())
	}
}

// BroadcastClientMessage send client message (not chat) to all clients
func (s *Server) BroadcastClientMessage(content string) {
	m := newClientMessage(content)
	s.BroadcastMessage(m)
}

// ChatMessage send a message from one player to another
// TODO: remove receiver player from roster on error?
func (s *Server) ChatMessage(to uuid.UUID, from string, msgText string) {
	conn := ROSTER.PlayerConn(to)
	m := newChatMessage(from, msgText)
	err := netSendMessage(conn, m)
	if err != nil {
		log.Printf("CHAT error sending to %s %s", conn.RemoteAddr().String(), err.Error())
	}
}

// RemovePlayer handles player leave events
func (s *Server) RemovePlayer(puuid uuid.UUID) {
	pname := ROSTER.PlayerName(puuid)
	if ROSTER.RemovePlayer(puuid) {
		msgText := fmt.Sprintf("Player %s left", pname)
		log.Printf(msgText)

		// chat message
		m := newBroadcastMessage(msgText)
		s.BroadcastMessage(m)

		// client message
		cm := newLeaveMessage(pname)
		s.BroadcastMessage(cm)

		// new scoreboard
		s.SendScores()
	}
}
