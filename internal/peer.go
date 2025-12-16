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

var peerCounter uint = 0

type PeerManager struct {
	peers []Peer
	mu    sync.RWMutex
	conn  *net.UDPConn

	// Callback for updating UI
	OnPeerListUpdated func([]Peer)
}

type Peer struct {
	ID       uint
	Addr     *net.UDPAddr
	Alias    string
	Status   Status
	LastSeen time.Time
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
	peerCounter++
	return &Peer{
		ID:       peerCounter,
		Addr:     address,
		Alias:    alias,
		Status:   status,
		LastSeen: time.Now(),
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
	pm.notifyUI()

	return nil
}

// Sends a message to a single peer using the global send socket
func (p Peer) SendMessage(conn *net.UDPConn, payload []byte) error {
	var err error = nil

	if p.Status == Active {
		_, err = conn.WriteToUDP(payload, p.Addr)
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
		if peer.Addr.IP.Equal(addr.IP) {
			alias = peer.Alias
		}
	}

	return alias
}

// Updates the alias on the specified peer
func (pm *PeerManager) SetAlias(Id uint, updatedAlias string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	for i, p := range pm.peers {
		if p.ID == Id {
			pm.peers[i].Alias = updatedAlias
		}
	}

	pm.notifyUI()
}

// Updates the lastSeen field of all known peers
// TODO: Check if i need this at some point?
func (pm *PeerManager) UpdateLastSeenAll() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	for _, peer := range pm.peers {
		peer.LastSeen = time.Now()
		peer.Status = Active
	}
}

// Update one specified peer if this peer is known. This is needed because one the whole peers list is known at some points
// of the application and the addr of the peer you want to update
func (pm *PeerManager) UpdatePeer(addr *net.UDPAddr) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	for i, peer := range pm.peers {
		log.Println("IP:", peer.Addr.IP)
		if peer.Addr.IP.Equal(addr.IP) {
			pm.peers[i].LastSeen = time.Now()
			pm.peers[i].Status = Active
		}
	}
}

func (pm *PeerManager) StartHeartbeatLoop() {
	ticker := time.NewTicker(5 * time.Second)

	go func() {
		for range ticker.C {
			pm.mu.Lock()

			// Remove inactive peers
			changed := pm.removeInactive()

			// Send ping to all known active peers
			pingMsg := CreatePing()
			err := pm.broadcastInternal(pingMsg)
			if err != nil {
				log.Fatal(err)
			}
			pm.mu.Unlock()

			// Update UI if we dropped a peer
			if changed {
				pm.notifyUI()
			}
		}
	}()
}

func (pm *PeerManager) removeInactive() bool {
	activePeers := (pm.peers)[:0]
	changed := false
	for _, peer := range pm.peers {
		if time.Since(peer.LastSeen) < timeout {
			activePeers = append(activePeers, peer)
		} else {
			log.Printf("Dropped: %s cause %s\n", peer.Addr.String(), time.Since(peer.LastSeen).String())
			changed = true
		}
	}

	pm.peers = activePeers
	return changed
}

func (pm *PeerManager) CloseSocket() {
	pm.conn.Close()
}

func (pm *PeerManager) ReadConn(buffer []byte) (int, *net.UDPAddr, error) {
	size, addr, err := pm.conn.ReadFromUDP(buffer)

	return size, addr, err
}

// Checks if a peer is known. Adds the peer to the list of active known peers if not known yet. Does nothing if the peer is known.
func (pm *PeerManager) CheckIfPeerIsKnown(addr *net.UDPAddr) bool {
	isKnown := false
	pm.mu.RLock()

	for _, peer := range pm.peers {
		if peer.Addr.IP.Equal(addr.IP) {
			isKnown = true
		}
	}

	pm.mu.RUnlock()

	if !isKnown {
		err := pm.Add(addr.String(), "")
		if err != nil {
			log.Println("Could not add new Peer:", err.Error())
		}
	}

	return isKnown
}

// Sends a ping via broadcast to everyone in the network. When they answer with pong we have found a new client
// and it will be added to the known active peers
func (pm *PeerManager) ScanNetwork() {
	msg := CreatePing()
	msg = append(msg, 0xAC, 0xAB) // This will be used to identify my own message so i dont add myself to peers
	target := "255.255.255.255:9000"

	addr, err := net.ResolveUDPAddr("udp", target)
	if err != nil {
		log.Fatal(err)
	}

	pm.conn.WriteToUDP(msg, addr)

}

func (pm *PeerManager) notifyUI() {
	if pm.OnPeerListUpdated != nil {
		// We create a copy here for preventing race conditions
		// We dont need a mutex lock though, because when this is called we will already be locked.
		peersCopy := make([]Peer, len(pm.peers))
		copy(peersCopy, pm.peers)

		go pm.OnPeerListUpdated(peersCopy)
	}
}
