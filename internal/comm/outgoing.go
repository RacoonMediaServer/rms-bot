package comm

import "github.com/RacoonMediaServer/rms-packages/pkg/communication"

// OutgoingMessage is a message from specific client to user
type OutgoingMessage struct {
	// Client Token
	Token string

	// Message is a message content
	Message *communication.BotMessage
}
