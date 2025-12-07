package internal

import "slices"

const (
	MsgTypeHeartbeat = 0x00
	MsgTypeUnicast   = 0x01
	MsgTypeBroadcast = 0x02
)

func HandleMesage(message []byte) (int, string) {
	var msgType int
	switch message[0] {
	case 0x00:
		msgType = MsgTypeHeartbeat
		// TODO: Handle heartbeat
	case 0x01:
		msgType = MsgTypeUnicast
		// TODO: Handle Unicast
	case 0x02:
		msgType = MsgTypeBroadcast
		// TODO: Handle Broadcast
	}

	return msgType, ""
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

func CreateUnicast(message string) []byte {
	header := []byte{0x01}
	content := []byte(message)
	payload := slices.Concat(header, content)
	return payload
}

func CreateBroadcast(message string) []byte {
	header := []byte{0x02}
	content := []byte(message)
	payload := slices.Concat(header, content)
	return payload
}

func HandleHeartbeat(message []byte) {

}
