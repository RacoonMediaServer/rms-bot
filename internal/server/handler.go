package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	rms_users "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-users"
	"github.com/gorilla/websocket"
	"go-micro.dev/v4/logger"
)

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 10 * time.Second,
}

type authResult struct {
	userId  string
	token   string
	selfReg bool
}

func (s *Server) authorize(ctx context.Context, token string) (authResult, error) {
	if token == "" {
		if s.selfRegistration {
			return s.performSelfRegistration(ctx)
		}
		return authResult{}, errors.New("invalid empty token")
	}

	req := rms_users.CheckPermissionsRequest{Token: token, Perms: []rms_users.Permissions{rms_users.Permissions_ConnectingToTheBot}}
	resp, err := s.f.NewUsers().CheckPermissions(ctx, &req)
	if err != nil {
		return authResult{}, err
	}

	if !resp.Allowed {
		return authResult{}, errors.New("access denied")
	}

	result := authResult{
		userId:  resp.UserId,
		token:   token,
		selfReg: false,
	}

	return result, nil
}

func (s *Server) performSelfRegistration(ctx context.Context) (authResult, error) {
	req := rms_users.User{
		Perms: []rms_users.Permissions{rms_users.Permissions_ConnectingToTheBot},
	}
	resp, err := s.f.NewUsers().RegisterUser(ctx, &req)
	if err != nil {
		return authResult{}, err
	}

	result := authResult{
		userId:  resp.UserId,
		token:   resp.Token,
		selfReg: true,
	}
	return result, nil
}

func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	l := s.l.Fields(map[string]interface{}{"addr": r.RemoteAddr})
	for key, val := range r.Header {
		l.Logf(logger.InfoLevel, "Got header %s = %+v", key, val)
	}
	token := r.Header.Get("X-Token")
	result, err := s.authorize(r.Context(), token)
	if err != nil {
		l.Logf(logger.ErrorLevel, "Forbidden: %s", err)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		l.Logf(logger.ErrorLevel, "Upgrade connection failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sess := newSession(s.l, conn, result.userId, s.ch)
	defer sess.close()

	s.mu.Lock()
	if existing, ok := s.sessions[result.userId]; ok {
		existing.drop()
	}
	s.sessions[result.userId] = sess
	s.mu.Unlock()

	sess.run(r.Context(), result)

	s.mu.Lock()
	if existing, ok := s.sessions[result.userId]; ok {
		if existing == sess {
			delete(s.sessions, result.userId)
		}
	}
	s.mu.Unlock()
}
