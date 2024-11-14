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

func (e *endpoint) authorize(ctx context.Context, token string) (authResult, error) {
	if token == "" {
		if e.selfReg {
			return e.performSelfRegistration(ctx)
		}
		return authResult{}, errors.New("invalid empty token")
	}

	req := rms_users.CheckPermissionsRequest{Token: token, Perms: []rms_users.Permissions{rms_users.Permissions_ConnectingToTheBot}}
	resp, err := e.f.NewUsers().CheckPermissions(ctx, &req)
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

func (e *endpoint) performSelfRegistration(ctx context.Context) (authResult, error) {
	req := rms_users.User{
		Perms: []rms_users.Permissions{rms_users.Permissions_ConnectingToTheBot},
	}
	resp, err := e.f.NewUsers().RegisterUser(ctx, &req)
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

func (e *endpoint) handler(w http.ResponseWriter, r *http.Request) {
	l := e.l.Fields(map[string]interface{}{"addr": r.RemoteAddr})
	for key, val := range r.Header {
		l.Logf(logger.InfoLevel, "Got header %s = %+v", key, val)
	}
	token := r.Header.Get("X-Token")
	result, err := e.authorize(r.Context(), token)
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

	sess := newSession(e.l, conn, result.userId, e.ch)
	defer sess.close()

	e.mu.Lock()
	if existing, ok := e.sessions[result.userId]; ok {
		existing.drop()
	}
	e.sessions[result.userId] = sess
	e.mu.Unlock()

	sess.run(r.Context(), result)

	e.mu.Lock()
	if existing, ok := e.sessions[result.userId]; ok {
		if existing == sess {
			delete(e.sessions, result.userId)
		}
	}
	e.mu.Unlock()
}
