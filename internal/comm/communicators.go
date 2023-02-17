package comm

import "errors"

// ErrDeviceIsNotConnected raises when server has no session with the token
var ErrDeviceIsNotConnected = errors.New("device is not connected")

// DeviceCommunicator sends IncomingMessage's to devices and receives OutgoingMessage's from devices
type DeviceCommunicator interface {
	Send(message IncomingMessage) error
	OutgoingChannel() <-chan OutgoingMessage
}
