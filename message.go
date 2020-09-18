package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	uuid "github.com/satori/go.uuid"
)

// MessageType used below
type MessageType int

// Types of messages
// TODO: keep this in sync with client code
const (
	CONNECT    MessageType = iota + 1 // 1
	DISCONNECT                        // 2
	PING                              // 3
	JUMP                              // 4
	MOVE                              // 5
	FIRE                              // 6
	HIT                               // 7
	JOIN                              // 8
	LEAVE                             // 9
	SCOREBOARD                        // 10
	BROADCAST                         // 11
	CHAT                              // 12
	CLIENT                            // 13
	KILL                              // 14
)

func (mt MessageType) String() string {
	switch mt {
	case CONNECT:
		return "CONNECT"
	case DISCONNECT:
		return "DISCONNECT"
	case PING:
		return "PING"
	case JUMP:
		return "JUMP"
	case MOVE:
		return "MOVE"
	case FIRE:
		return "FIRE"
	case HIT:
		return "HIT"
	case JOIN:
		return "JOIN"
	case LEAVE:
		return "LEAVE"
	case SCOREBOARD:
		return "SCOREBOARD"
	case BROADCAST:
		return "BROADCAST"
	case CHAT:
		return "CHAT"
	case CLIENT:
		return "CLIENT"
	case KILL:
		return "KILL"
	default:
		return "UNKNOWN"
	}
}

// Message is the format of messages to/from clients
type Message struct {
	Damage     float32          `json:"damage,omitempty"`     // damage dealt to HitPlayer
	HitPlayer  string           `json:"hit_player,omitempty"` // player hit
	Name       string           `json:"name,omitempty"`       // player name
	Sender     uuid.UUID        `json:"-"`                    // internal use only
	Type       MessageType      `json:"type"`                 // message type
	Content    string           `json:"content,omitempty"`    // message content
	MessageTo  string           `json:"message_to,omitempty"` // player or team name message is sent to
	Scoreboard map[string]int32 `json:"scoreboard,omitempty"` // scoreboard contents
	Nonce      string           `json:"nonce,omitempty"`      // just a nonce (use for ping?)
	Result     string           `json:"result,omitempty"`     // to indicate a result of some sort
	Attacker   string           `json:"attacker,omitempty"`   // name of attacker
}

func parseClientMessage(conn net.Conn, raw []byte) {
	var msg Message
	err := json.Unmarshal(raw, &msg)
	if err != nil {
		log.Printf("CLNT error parsing client message: %s", err.Error())
		return
	}

	msg.Sender = ROSTER.PlayerLookup("", conn.RemoteAddr())
	ProcessMessage(conn, msg)
}

func newScoreboardMessage() Message {
	return Message{
		Type:       SCOREBOARD,
		Scoreboard: ROSTER.Scoreboard(),
	}
}

func newBroadcastMessage(content string) Message {
	return Message{
		Type:    BROADCAST,
		Content: content,
	}
}

func newClientMessage(content string) Message {
	return Message{
		Type:    CLIENT,
		Content: content,
	}
}

func newChatMessage(from, content string) Message {
	return Message{
		Type: CHAT,
		Name: from,
	}
}

func newJoinMessage(puuid, name, result string) Message {
	return Message{
		Type:    JOIN,
		Name:    name,
		Content: puuid,
		Result:  result,
	}
}

func newKillMessage(attacker string) Message {
	return Message{
		Type:     KILL,
		Attacker: attacker,
	}
}

func newLeaveMessage(name string) Message {
	return Message{
		Type: LEAVE,
		Name: name,
	}
}

// ProcessMessage handles all incoming client messages
func ProcessMessage(conn net.Conn, msg Message) {
	log.Printf("Message received (%s) from %s", msg.Type.String(), conn.RemoteAddr().String())
	// unauthenticated messages
	switch mt := msg.Type; mt {
	case CONNECT:
		log.Printf("Client connected %s", conn.RemoteAddr().String())
		return
	case DISCONNECT:
		fmt.Println("Client disconnected")
		return
	case PING:
		fmt.Println("Client ping")
		return
	case JOIN:
		p := newPlayer(msg.Name, conn)
		SVR.PlayerJoin(p)
		return
	}

	// drop unauthenticated messages
	if msg.Sender == uuid.Nil {
		log.Print("MSG dropped due to unknown sender")
		return
	}

	pn := ROSTER.PlayerName(msg.Sender)

	switch mt := msg.Type; mt {
	case JUMP:
		log.Printf("Player %s jump", pn)
	case MOVE:
		log.Printf("Player %s move", pn)
	case FIRE:
		log.Printf("Player %s fireed", pn)
	case HIT:
		SVR.PlayerHit(msg.Name, msg.Sender, 50.0)
	case LEAVE:
		SVR.RemovePlayer(msg.Sender)
	case SCOREBOARD:
		SVR.SendScores()
	case BROADCAST:
		fmt.Println("broadcast")
	case CHAT:
		fmt.Println("chat")
	default:
		fmt.Printf("unknown message type %d.\n", mt)
	}
	// authenticated messages
}
