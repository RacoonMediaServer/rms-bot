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
	l     logger.Logger
	conn  *websocket.Conn
	token string
	user  chan *communication.UserMessage
	out   chan<- comm.OutgoingMessage
}

func newSession(l logger.Logger, conn *websocket.Conn, token string, out chan<- comm.OutgoingMessage) *session {
	return &session{
		l:     l.Fields(map[string]interface{}{"token": token}),
		conn:  conn,
		token: token,
		user:  make(chan *communication.UserMessage, maxMessageQueueSize),
		out:   out,
	}
}

func (s *session) run(ctx context.Context) {
	s.l.Log(logger.InfoLevel, "Established")

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
			s.l.Logf(logger.ErrorLevel, "pick message failed: %s", s.token, err)
			return
		}
		msg := &communication.BotMessage{}
		if err = proto.Unmarshal(buf, msg); err != nil {
			s.l.Logf(logger.ErrorLevel, "deserialize incoming message failed: %s", err)
			return
		}
		s.l.Logf(logger.DebugLevel, "message from client received: %s", msg)
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
				s.l.Logf(logger.ErrorLevel, "serialize message failed: %s", err)
				continue
			}
			if err = s.conn.WriteMessage(websocket.BinaryMessage, buf); err != nil {
				s.l.Logf(logger.ErrorLevel, "write message failed: %s", err)
				continue
			}
			s.l.Logf(logger.DebugLevel, "message sent to client: %s", msg)
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
