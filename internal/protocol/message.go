package protocol

const (
	MessageTypeMessage = "MESSAGE"
	MessageTypePing    = "PING"
	MessageTypePong    = "PONG"
)

type Message struct {
	Type string `json:"type"`
	From string `json:"from"`
	Data string `json:"data"`
}
