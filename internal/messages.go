package internal

import (
	"log"
	"net"
)

const (
	MsgTypeHeartbeat = 0x00
	MsgTypeUnicast   = 0x01
	MsgTypeBroadcast = 0x02
	EmptyPacket      = 0x63 // Hex 63 = 99 int
)

// HandelIncomingPacket handels all incoming packets. It decides what to do based on the message type. It needs:
// pm: the peerManager holding the socket and peer list
// addr: The address the messsage was sent from
// data: the byte slice of data received
//
// It returns the msgType, the content of the message and any errors
func HandleIncomingPacket(pm *PeerManager, addr *net.UDPAddr, data []byte) (byte, string, error) {
	if len(data) < 1 {
		return EmptyPacket, "", nil // 99 indicating empty packet
	}

	// First update the peer lastSeen time, since the peer is still active if we receive a msg
	pm.UpdatePeer(addr)

	msgType := data[0]
	content := data[1:]

	switch msgType {
	case MsgTypeHeartbeat:
		if len(content) > 1 && content[1] == 0xAC && content[2] == 0xAB {
			log.Println("Reveived own message from broadcast")
			return EmptyPacket, "", nil
		}
		// Check if peer is known, add otherwise
		isKnown := pm.CheckIfPeerIsKnown(addr)
		if !isKnown {
			log.Println("New Peer added:", addr.String())
		}
		if content[0] == 0x00 {
			log.Printf("%s has sent a Ping:%s\n", addr.String(), string(content))
			pong := CreatePong()
			pm.conn.WriteToUDP(pong, addr)
		} else { // must be an incoming Pong (0x01)
			log.Printf("%s returned Pong :yey: %s\n", addr.String(), string(content))
		}
	case MsgTypeUnicast:
		return MsgTypeUnicast, string(content), nil
	case MsgTypeBroadcast:
		return MsgTypeBroadcast, string(content), nil
	}

	return msgType, string(content), nil
}

func ParseMessage(data []byte) (byte, []byte) {
	if len(data) == 0 {
		return 0, []byte{0}
	}

	return data[0], data[1:]
}

func PrepareMessage(msgType byte, content string) []byte {
	payload := []byte(content)
	return append([]byte{msgType}, payload...)
}

func CreatePing() []byte {
	return []byte{0x00, 0x00}
}

func CreatePong() []byte {
	return []byte{0x00, 0x01}
}
