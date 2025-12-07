package internal

import (
	"errors"
	"net"
)

type Status int

const (
	Active Status = iota
	Dead
	Unknown
)

type Peer struct {
	addr   *net.UDPAddr
	alias  string
	status Status
}

type Peers []Peer

func NewPeer(address *net.UDPAddr, alias string, status Status) *Peer {
	return &Peer{
		addr:   address,
		alias:  alias,
		status: status,
	}
}

func (peers *Peers) Add(peer string, alias string) error {
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

	newPeer := NewPeer(addr, alias, Active)
	*peers = append(*peers, *newPeer)

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
func (p *Peers) Broadcast(conn *net.UDPConn, payload []byte) error {
	var err error = nil
	for _, peer := range *p {
		err = peer.SendMessage(conn, payload)
	}

	return err
}

// Returns the alias matching the given address if the peer is already known to the system
// Returns the address of the peer otherwise
func (p *Peers) GetAlias(addr *net.UDPAddr) string {
	alias := addr.String()
	for _, peer := range *p {
		if peer.addr.IP.Equal(addr.IP) {
			alias = peer.alias
		}
	}

	return alias
}
