package internal

import "encoding/json"

type Message struct {
	Payload json.RawMessage `json: "payload"`
}
