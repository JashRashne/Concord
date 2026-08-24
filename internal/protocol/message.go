package protocol

type Message struct {
	Type string `json:"type"`
	From string `json:"from"`
	Data string `json:"data"`
}
