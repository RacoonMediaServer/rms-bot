package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/RacoonMediaServer/rms-bot-server/internal/db"
	"github.com/RacoonMediaServer/rms-packages/pkg/service/servicemgr"
	"go-micro.dev/v4/logger"
	"net"
	"net/http"
	"sync"
	"time"
)

type Server interface {
	ListenAndServe(host string, port int) error
	Shutdown()
	Wait()
}

type server struct {
	db       db.Database
	f        servicemgr.ServiceFactory
	s        http.Server
	wg       sync.WaitGroup
	sessions sync.Map
}

func New(database db.Database, f servicemgr.ServiceFactory) Server {
	s := &server{
		db: database,
		f:  f,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/bot", s.handler)

	s.s.Handler = mux
	return s
}

func (s *server) ListenAndServe(host string, port int) error {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("bind server address failed: %w", err)
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err = s.s.Serve(l); !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal(err)
		}
	}()

	return nil
}

func (s *server) Shutdown() {
	const shutdownTimeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	_ = s.s.Shutdown(ctx)
	s.wg.Wait()
}

func (s *server) Wait() {
	s.wg.Wait()
}
