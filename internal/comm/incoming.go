package comm

import "github.com/RacoonMediaServer/rms-packages/pkg/communication"

// IncomingMessage is a message from user to specific device
type IncomingMessage struct {
	DeviceID string

	// Message represents message content
	Message *communication.UserMessage
}
