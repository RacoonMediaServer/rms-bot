package comm

import "github.com/RacoonMediaServer/rms-packages/pkg/communication"

// IncomingMessage is a message from user to specific client
type IncomingMessage struct {
	// Client Token
	Token string

	// Message represents message content
	Message *communication.UserMessage
}
