package protocol

const (
	MessageTypeMessage = "MESSAGE"
	MessageTypePing    = "PING"
	MessageTypePong    = "PONG"
	MessageTypeSet     = "SET"
)

type Message struct {
	Type  string `json:"type"`
	From  string `json:"from"`
	Data  string `json:"data,omitempty"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}
