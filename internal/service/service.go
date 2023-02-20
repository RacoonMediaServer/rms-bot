package service

import (
	"context"
	rms_bot_server "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-bot-server"
	"go-micro.dev/v4/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server interface {
	DropSession(token string)
}

type Database interface {
	RemoveDevice(deviceID string) error
}

type service struct {
	l      logger.Logger
	server Server
	db     Database
}

func (s service) DropSession(ctx context.Context, request *rms_bot_server.DropSessionRequest, empty *emptypb.Empty) error {
	s.l.Logf(logger.InfoLevel, "Drop session %s", request.Token)
	s.server.DropSession(request.Token)
	if err := s.db.RemoveDevice(request.Token); err != nil {
		s.l.Logf(logger.WarnLevel, "Cannot drop device linkage: %s", err)
	}
	return nil
}

func New(server Server, db Database) rms_bot_server.RmsBotServerHandler {
	return &service{
		server: server,
		db:     db,
		l:      logger.DefaultLogger.Fields(map[string]interface{}{"from": "service"}),
	}
}
