package main

import (
	"fmt"
	"math/rand"
	"net"
	"sync"

	uuid "github.com/satori/go.uuid"
)

// Roster is our list of players, along with their kills and deaths
type Roster struct {
	lock    sync.RWMutex
	players map[uuid.UUID]Player
	kills   map[uuid.UUID]int32
	deaths  map[uuid.UUID]int32
}

// newRoster generates a new, empty roster
func newRoster() Roster {
	return Roster{
		players: make(map[uuid.UUID]Player),
		kills:   make(map[uuid.UUID]int32),
		deaths:  make(map[uuid.UUID]int32),
	}
}

// AddPlayer does just that
// TODO: ensure another player doesn't have the same name, if so add integers to end of this player's name until it is unique
func (r *Roster) AddPlayer(p Player) bool {
	// Ensure we have a unique uuid
	if !r.PlayerIDUnique(p.ID()) {
		return false
	}

	// Ensure we have a unique remote address (and port)
	if !r.PlayerConnUnique(p.Connection()) {
		return false
	}

	// Ensure we have a unique name
	// TODO: send this back to the client
	for !r.PlayerNameUnique(p.Name()) {
		p.SetName(fmt.Sprintf("%s%d", p.Name(), rand.Intn(10)))
	}

	// Add user object to our roster
	r.lock.Lock()
	r.players[p.ID()] = p
	r.deaths[p.ID()] = 0
	r.kills[p.ID()] = 0
	r.lock.Unlock()
	return true
}

// RemovePlayer does just that
func (r *Roster) RemovePlayer(puuid uuid.UUID) bool {
	var success bool
	r.lock.Lock()
	if p, ok := r.players[puuid]; ok {
		p.Connection().Close()
		delete(r.players, puuid)
		delete(r.kills, puuid)
		delete(r.deaths, puuid)
		success = true
	}
	r.lock.Unlock()
	return success
}

// PlayerKilled updates scoreboard on player kills
func (r *Roster) PlayerKilled(vUUID, aUUID uuid.UUID) {
	r.lock.Lock()
	if _, ok := r.kills[aUUID]; ok {
		r.kills[aUUID]++
	} else {
		r.kills[aUUID] = 1
	}
	r.deaths[vUUID]++
	r.lock.Unlock()
}

// PlayerKills returns the number of kills of a player
func (r *Roster) PlayerKills(puuid uuid.UUID) int32 {
	var kills int32
	r.lock.RLock()
	if k, ok := r.kills[puuid]; ok {
		kills = k
	}
	r.lock.RUnlock()
	return kills
}

// PlayerDeaths returns the number of deaths of a player
func (r *Roster) PlayerDeaths(puuid uuid.UUID) int32 {
	var deaths int32
	r.lock.RLock()
	if d, ok := r.deaths[puuid]; ok {
		deaths = d
	}
	r.lock.RUnlock()
	return deaths
}

// PlayerLookup retrievs the player UUID given address or name
func (r *Roster) PlayerLookup(name string, addr net.Addr) uuid.UUID {
	var pu uuid.UUID

	r.lock.RLock()
	for puuid, player := range r.players {
		if name != "" && name == player.Name() {
			pu = puuid
			break
		} else if player.Connection() != nil && addr != nil && player.Connection().RemoteAddr().String() == addr.String() {
			pu = puuid
			break
		}
	}
	r.lock.RUnlock()
	return pu
}

// Scoreboard generates a list of player kills
func (r *Roster) Scoreboard() map[string]int32 {
	var scores = make(map[string]int32)
	r.lock.RLock()
	for puuid, kills := range r.kills {
		pname := ROSTER.PlayerName(puuid)
		scores[pname] = kills
	}
	r.lock.RUnlock()
	return scores
}

// PlayerName looks up a player's name
func (r *Roster) PlayerName(puuid uuid.UUID) (name string) {
	r.lock.RLock()
	if p, ok := r.players[puuid]; ok {
		name = p.Name()
	}
	r.lock.RUnlock()
	return
}

// PlayerHealth looks up a player's health
func (r *Roster) PlayerHealth(puuid uuid.UUID) (health float32) {
	r.lock.RLock()
	if p, ok := r.players[puuid]; ok {
		health = p.Health()
	}
	r.lock.RUnlock()
	return
}

// PlayerDamage subtracts damage form a player's health
func (r *Roster) PlayerDamage(puuid uuid.UUID, damage float32) (health float32) {
	r.lock.Lock()
	p := r.players[puuid]
	prettyPrint(p)
	p.TakeDamage(damage)
	health = p.Health()
	r.players[puuid] = p
	r.lock.Unlock()
	return
}

// RespawnPlayer resets player health and status
func (r *Roster) RespawnPlayer(puuid uuid.UUID) {
	r.lock.Lock()
	if p, ok := r.players[puuid]; ok {
		p.Respawn(100.00)
		r.players[puuid] = p
	}
	r.lock.Unlock()
}

// ResetPlayer replenishes player health and zeroes out player scoreboard stats
func (r *Roster) ResetPlayer(puuid uuid.UUID) {
	r.lock.Lock()
	if p, ok := r.players[puuid]; ok {
		p.Respawn(100.00)
		r.players[puuid] = p
	}
	r.kills[puuid] = 0
	r.deaths[puuid] = 0
	r.lock.Unlock()
}

// Reset zeroes out all roster counters
func (r *Roster) Reset() {
	r.lock.Lock()
	r.kills = make(map[uuid.UUID]int32)
	r.deaths = make(map[uuid.UUID]int32)
	r.lock.Unlock()
	for puuid := range r.players {
		r.ResetPlayer(puuid)
	}
}

// PlayerNameUnique determines if a player's name is unique
func (r *Roster) PlayerNameUnique(name string) bool {
	var unique = true
	r.lock.RLock()
	for _, p := range r.players {
		if name == p.Name() {
			unique = false
			break
		}
	}
	r.lock.RUnlock()
	return unique
}

// PlayerConnUnique determines if a player's connection (remote address and port) is unique
func (r *Roster) PlayerConnUnique(conn net.Conn) bool {
	var unique = true
	r.lock.RLock()
	for _, p := range r.players {
		if conn.RemoteAddr().String() == p.Connection().RemoteAddr().String() {
			unique = false
			break
		}
	}
	r.lock.RUnlock()
	return unique
}

// PlayerIDUnique determines if a player's UUID is unique
func (r *Roster) PlayerIDUnique(puuid uuid.UUID) bool {
	var unique = true
	r.lock.RLock()
	for pID := range r.players {
		if puuid == pID {
			unique = false
			break
		}
	}
	r.lock.RUnlock()
	return unique
}

// PlayerConn retrieves a player's connection
func (r *Roster) PlayerConn(puuid uuid.UUID) net.Conn {
	r.lock.RLock()
	if p, ok := r.players[puuid]; ok {
		return p.Connection()
	}
	r.lock.RUnlock()
	return nil
}

// PlayerConnAll retrieves all player connections
func (r *Roster) PlayerConnAll() []net.Conn {
	r.lock.RLock()
	var conns []net.Conn
	for _, p := range r.players {
		conns = append(conns, p.Connection())
	}
	r.lock.RUnlock()
	return conns
}

// PlayerCount returns the number of players
func (r *Roster) PlayerCount() int {
	return len(r.players)
}
