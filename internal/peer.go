package internal

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

type Status int

const timeout = 30 * time.Second
const (
	Active Status = iota
	Dead
	Unknown
)

type PeerManager struct {
	peers []Peer
	mu    sync.RWMutex
	conn  *net.UDPConn
}

type Peer struct {
	addr     *net.UDPAddr
	alias    string
	status   Status
	lastSeen time.Time
}

type Peers []Peer

func NewPeerManager(port string) *PeerManager {
	addr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	return &PeerManager{
		peers: []Peer{},
		mu:    sync.RWMutex{},
		conn:  conn,
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

// Sends the message to all known and active peers using the pm socket.
// This is thread safe and can be safely called externaly
func (pm *PeerManager) Broadcast(payload []byte) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	err := pm.broadcastInternal(payload)

	return err
}

// Sends the message to all known and active peers.
// This is NOT thread safe and needs to be protected manually
func (pm *PeerManager) broadcastInternal(payload []byte) error {
	var err error
	for _, peer := range pm.peers {
		err = peer.SendMessage(pm.conn, payload)
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
		peer.status = Active
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
			peer.status = Active
		}
	}
}

func (pm *PeerManager) StartHeartbeatLoop() {
	ticker := time.NewTicker(5 * time.Second)

	go func() {
		for range ticker.C {
			pm.mu.Lock()

			// Remove inactive peers
			pm.removeInactive()

			// Send ping to all known active peers
			pingMsg := CreatePing()
			err := pm.broadcastInternal(pingMsg)
			if err != nil {
				log.Fatal(err)
			}

			pm.mu.Unlock()
		}
	}()
}

func (pm *PeerManager) removeInactive() {
	activePeers := (pm.peers)[:0]
	for _, peer := range pm.peers {
		if time.Since(peer.lastSeen) < timeout {
			activePeers = append(activePeers, peer)
		}
	}

	pm.peers = activePeers
}

func (pm *PeerManager) CloseSocket() {
	pm.conn.Close()
}

func (pm *PeerManager) ReadConn(buffer []byte) (int, *net.UDPAddr, error) {
	size, addr, err := pm.conn.ReadFromUDP(buffer)

	return size, addr, err
}
