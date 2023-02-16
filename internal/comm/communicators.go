package comm

// ClientCommunicator sends IncomingMessage's to clients and receives OutgoingMessage's from clients
type ClientCommunicator interface {
	Send(message IncomingMessage) error
	OutgoingChannel() <-chan OutgoingMessage
}
