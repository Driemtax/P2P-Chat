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

func (peers *Peers) Add(peer string) error {
	if peer == "" {
		return errors.New("Cannot add empty peer...")
	}

	addr, err := net.ResolveUDPAddr("udp", peer)
	if err != nil {
		return err
	}

	newPeer := NewPeer(addr, "", Active)
	*peers = append(*peers, *newPeer)

	return nil
}

// Sends a message to a single peer using the global send socket
func (p Peer) SendMessage(conn *net.UDPConn, msg string) error {
	var err error = nil

	if p.status == Active {
		_, err = conn.WriteToUDP([]byte(msg), p.addr)
	}

	return err
}

// Sends the message to all known and active peers using the specified socket.
func (p *Peers) Broadcast(conn *net.UDPConn, msg string) error {
	var err error = nil
	for _, peer := range *p {
		err = peer.SendMessage(conn, msg)
	}

	return err
}
