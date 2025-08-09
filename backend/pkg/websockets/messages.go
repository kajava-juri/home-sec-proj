package websockets

type ClientMessage struct {
	Action string `json:"action"`
	Topics []string `json:"topics,omitempty"`
}