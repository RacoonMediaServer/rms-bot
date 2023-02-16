package server

import (
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

func (s *server) authorize(r *http.Request) (bool, error) {
	token := r.Header.Get("X-Token")
	if token == "" {
		logger.Warnf("[%s] Unauthorized access attempt", r.RemoteAddr)
		return false, nil
	}

	resp, err := s.f.NewUsers().GetPermissions(r.Context(), &rms_users.GetPermissionsRequest{Token: token})
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

func (s *server) handler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("X-Token")
	if ok, err := s.authorize(r); !ok {
		if err != nil {
			logger.Errorf("[%s]: Authorization failed: %s", r.RemoteAddr, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logger.Warnf("[%s]: Forbidden", r.RemoteAddr)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sess := newSession(conn)
	defer sess.close()

	s.sessions.Store(token, sess)
	sess.run()
	s.sessions.Delete(token)
}
