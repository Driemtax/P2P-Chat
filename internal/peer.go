package internal

import (
	"errors"
	"net"
	"sync"
	"time"
)

type Status int

const (
	Active Status = iota
	Dead
	Unknown
)

type PeerManager struct {
	peers []Peer
	mu    sync.RWMutex
}

type Peer struct {
	addr     *net.UDPAddr
	alias    string
	status   Status
	lastSeen time.Time
}

type Peers []Peer

func NewPeerManager() *PeerManager {
	return &PeerManager{
		peers: []Peer{},
		mu:    sync.RWMutex{},
	}
}

func NewPeer(address *net.UDPAddr, alias string, status Status) *Peer {
	return &Peer{
		addr:     address,
		alias:    alias,
		status:   status,
		lastSeen: time.Now(),
	}
}

func (pm *PeerManager) Add(peer string, alias string) error {
	if peer == "" {
		return errors.New("Cannot add empty peer...")
	}

	addr, err := net.ResolveUDPAddr("udp", peer)
	if err != nil {
		return err
	}

	if alias == "" {
		alias = peer
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()
	newPeer := NewPeer(addr, alias, Active)
	pm.peers = append(pm.peers, *newPeer)

	return nil
}

// Sends a message to a single peer using the global send socket
func (p Peer) SendMessage(conn *net.UDPConn, payload []byte) error {
	var err error = nil

	if p.status == Active {
		_, err = conn.WriteToUDP(payload, p.addr)
	}

	return err
}

// Sends the message to all known and active peers using the specified socket.
func (pm *PeerManager) Broadcast(conn *net.UDPConn, payload []byte) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	var err error = nil
	for _, peer := range pm.peers {
		err = peer.SendMessage(conn, payload)
	}

	return err
}

// Returns the alias matching the given address if the peer is already known to the system
// Returns the address of the peer otherwise
func (pm *PeerManager) GetAlias(addr *net.UDPAddr) string {
	alias := addr.String()
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	for _, peer := range pm.peers {
		if peer.addr.IP.Equal(addr.IP) {
			alias = peer.alias
		}
	}

	return alias
}

// Updates the lastSeen field of all known peers
// TODO: Check if i need this at some point?
func (pm *PeerManager) UpdateLastSeenAll() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	for _, peer := range pm.peers {
		peer.lastSeen = time.Now()
	}
}

// Update one specified peer if this peer is known. This is needed because one the whole peers list is known at some points
// of the application and the addr of the peer you want to update
func (pm *PeerManager) UpdatePeer(addr *net.UDPAddr) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	for _, peer := range pm.peers {
		if peer.addr.IP.Equal(addr.IP) {
			peer.lastSeen = time.Now()
		}
	}
}

func (pm *PeerManager) CheckAll(conn *net.UDPConn) {
	now := time.Now()
	// TODO: Subtract 30 Seconds

	for _, peer := range pm.peers {
		// Peer needs to be declared dead and removed from peerList
		if peer.lastSeen.Before(now) {
			peer.status = Dead
		}
	}
}
