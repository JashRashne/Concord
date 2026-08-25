package protocol

const (
	MessageTypeMessage  = "MESSAGE"
	MessageTypePing     = "PING"
	MessageTypePong     = "PONG"
	MessageTypeSet      = "SET"
	MessageTypeGet      = "GET"
	MessageTypeValue    = "VALUE"
	MessageTypeNotFound = "NOT_FOUND"
)

type Message struct {
	Type  string `json:"type"`
	From  string `json:"from"`
	Data  string `json:"data,omitempty"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}
