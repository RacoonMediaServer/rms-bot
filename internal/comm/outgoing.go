package comm

import "github.com/RacoonMediaServer/rms-packages/pkg/communication"

// OutgoingMessage is a message from specific device to user
type OutgoingMessage struct {
	DeviceID string

	// Message is a message content
	Message *communication.BotMessage
}
