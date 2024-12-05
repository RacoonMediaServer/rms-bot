package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/RacoonMediaServer/rms-bot-server/internal/comm"
	"github.com/RacoonMediaServer/rms-packages/pkg/middleware"
	"github.com/RacoonMediaServer/rms-packages/pkg/service/servicemgr"
	"go-micro.dev/v4/logger"
)

const maxMessageQueueSize = 1000

type Server struct {
	l  logger.Logger
	s  http.Server
	wg sync.WaitGroup

	endpoints map[string]*endpoint
}

func New(f servicemgr.ServiceFactory, endpoints []comm.Endpoint) *Server {
	s := &Server{
		l:         logger.DefaultLogger.Fields(map[string]interface{}{"from": "server"}),
		endpoints: map[string]*endpoint{},
	}

	mux := http.NewServeMux()

	for _, endpoint := range endpoints {
		e := newEndpoint(s.l.Fields(map[string]interface{}{"endpoint": endpoint.ID}),
			f, endpoint.ID)
		s.endpoints[endpoint.ID] = e

		mux.HandleFunc("/bot/"+endpoint.ID, e.handler)
	}

	s.s.Handler = middleware.PanicHandler(middleware.UnauthorizedRequestsCountHandler(mux))
	return s
}

func (s *Server) GetEndpoint(id string) (comm.DeviceCommunicator, error) {
	e, ok := s.endpoints[id]
	if !ok {
		return nil, errors.New("unknown endpoint")
	}
	return e, nil
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

func (s *Server) DropSession(endpoint, user string) {
	e, ok := s.endpoints[endpoint]
	if !ok {
		return
	}
	e.dropSession(user)
}
