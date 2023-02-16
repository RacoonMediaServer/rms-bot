package server

import (
	"context"
	"github.com/RacoonMediaServer/rms-bot-server/internal/comm"
	"github.com/RacoonMediaServer/rms-packages/pkg/communication"
	"github.com/gorilla/websocket"
	"go-micro.dev/v4/logger"
	"google.golang.org/protobuf/proto"
	"sync"
)

type session struct {
	conn  *websocket.Conn
	token string
	user  chan *communication.UserMessage
	out   chan<- comm.OutgoingMessage
}

func newSession(conn *websocket.Conn, token string, out chan<- comm.OutgoingMessage) *session {
	return &session{
		conn:  conn,
		token: token,
		user:  make(chan *communication.UserMessage, maxMessageQueueSize),
		out:   out,
	}
}

func (s *session) run(ctx context.Context) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	defer wg.Wait()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		defer wg.Done()
		s.writeProcess(ctx)
	}()

	for {
		// читаем сообщения от клиентского устройства
		_, buf, err := s.conn.ReadMessage()
		if err != nil {
			logger.Errorf("[%s] pick message failed: %s", s.token, err)
			return
		}
		msg := &communication.BotMessage{}
		if err = proto.Unmarshal(buf, msg); err != nil {
			logger.Errorf("[%s] deserialize incoming message failed: %s", s.token, err)
			return
		}
		logger.Debugf("[%s] message from client received: %s", s.token, msg)
		s.out <- comm.OutgoingMessage{Token: s.token, Message: msg}
	}
}

func (s *session) writeProcess(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-s.user:
			buf, err := proto.Marshal(msg)
			if err != nil {
				logger.Errorf("[%s] serialize message failed: %s", s.token, err)
				continue
			}
			if err = s.conn.WriteMessage(websocket.BinaryMessage, buf); err != nil {
				logger.Errorf("[%s] write message failed: %s", s.token, err)
				continue
			}
			logger.Debugf("[%s] message sent to client: %s", s.token, msg)
		}
	}
}

func (s *session) send(msg *communication.UserMessage) {
	s.user <- msg
}

func (s *session) close() {
	_ = s.conn.Close()
	close(s.user)
}
