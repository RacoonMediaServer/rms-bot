package server

import (
	"sync"

	"github.com/RacoonMediaServer/rms-bot-server/internal/comm"
	"github.com/RacoonMediaServer/rms-packages/pkg/service/servicemgr"
	"go-micro.dev/v4/logger"
)

type endpoint struct {
	l       logger.Logger
	f       servicemgr.ServiceFactory
	selfReg bool
	ch      chan comm.OutgoingMessage

	mu       sync.RWMutex
	sessions map[string]*session
}

func newEndpoint(l logger.Logger, f servicemgr.ServiceFactory, selfRegistration bool) *endpoint {
	return &endpoint{
		l:        l,
		f:        f,
		sessions: make(map[string]*session),
		ch:       make(chan comm.OutgoingMessage, maxMessageQueueSize),
		selfReg:  selfRegistration,
	}
}

func (e *endpoint) OutgoingChannel() <-chan comm.OutgoingMessage {
	return e.ch
}

func (e *endpoint) Send(message comm.IncomingMessage) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	sess, ok := e.sessions[message.DeviceID]
	if !ok {
		return comm.ErrDeviceIsNotConnected
	}

	sess.send(message.Message)

	return nil
}

func (e *endpoint) dropSession(user string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	sess, ok := e.sessions[user]
	if !ok {
		return
	}

	sess.drop()
	delete(e.sessions, user)
}
