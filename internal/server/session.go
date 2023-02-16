package server

import "github.com/gorilla/websocket"

type session struct {
	conn *websocket.Conn
}

func newSession(conn *websocket.Conn) *session {
	return &session{conn: conn}
}

func (s *session) run() {

}

func (s *session) close() {
	_ = s.conn.Close()
}
