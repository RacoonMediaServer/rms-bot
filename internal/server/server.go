package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/RacoonMediaServer/rms-bot-server/internal/comm"
	"github.com/RacoonMediaServer/rms-packages/pkg/service/servicemgr"
	"go-micro.dev/v4/logger"
	"net"
	"net/http"
	"sync"
	"time"
)

var ErrClientIsNotConnected = errors.New("client is not connectes")

const maxMessageQueueSize = 1000

type Server struct {
	l  logger.Logger
	f  servicemgr.ServiceFactory
	s  http.Server
	wg sync.WaitGroup
	ch chan comm.OutgoingMessage

	mu       sync.RWMutex
	sessions map[string]*session
}

func (s *Server) OutgoingChannel() <-chan comm.OutgoingMessage {
	return s.ch
}

func (s *Server) Send(message comm.IncomingMessage) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sess, ok := s.sessions[message.Token]
	if !ok {
		return ErrClientIsNotConnected
	}

	sess.send(message.Message)

	return nil
}

func New(f servicemgr.ServiceFactory) *Server {
	s := &Server{
		l:        logger.DefaultLogger.Fields(map[string]interface{}{"from": "server"}),
		f:        f,
		sessions: make(map[string]*session),
		ch:       make(chan comm.OutgoingMessage, maxMessageQueueSize),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/bot", s.handler)

	s.s.Handler = mux
	return s
}

func (s *Server) ListenAndServe(host string, port int) error {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("bind Server address failed: %w", err)
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err = s.s.Serve(l); !errors.Is(err, http.ErrServerClosed) {
			s.l.Log(logger.ErrorLevel, err)
		}
	}()

	return nil
}

func (s *Server) Shutdown() {
	const shutdownTimeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	_ = s.s.Shutdown(ctx)
	s.wg.Wait()
}

func (s *Server) Wait() {
	s.wg.Wait()
}

func (s *Server) DropSession(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[token]
	if !ok {
		return
	}

	sess.close()
	delete(s.sessions, token)
}
