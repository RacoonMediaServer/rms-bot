package server

import (
	"context"
	"fmt"
	rms_users "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-users"
	"github.com/gorilla/websocket"
	"go-micro.dev/v4/logger"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 10 * time.Second,
}

func (s *Server) authorize(ctx context.Context, token string) (bool, error) {
	if token == "" {
		return false, nil
	}

	resp, err := s.f.NewUsers().GetPermissions(ctx, &rms_users.GetPermissionsRequest{Token: token})
	if err != nil {
		return false, fmt.Errorf("token validation failed: %w", err)
	}
	for _, perm := range resp.Perms {
		if perm == rms_users.Permissions_ConnectingToTheBot {
			return true, nil
		}
	}

	return false, nil
}

func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	l := s.l.Fields(map[string]interface{}{"addr": r.RemoteAddr})
	token := r.Header.Get("X-Token")
	if ok, err := s.authorize(r.Context(), token); !ok {
		if err != nil {
			l.Logf(logger.ErrorLevel, "Authorization failed: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		l.Log(logger.ErrorLevel, "Forbidden")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sess := newSession(s.l, conn, token, s.ch)
	defer sess.close()

	s.mu.Lock()
	s.sessions[token] = sess
	s.mu.Unlock()

	sess.run(r.Context())

	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}
