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

type service struct {
	l      logger.Logger
	server Server
}

func (s service) DropSession(ctx context.Context, request *rms_bot_server.DropSessionRequest, empty *emptypb.Empty) error {
	s.l.Logf(logger.InfoLevel, "Drop session %s", request.Token)
	s.server.DropSession(request.Token)
	return nil
}

func New(server Server) rms_bot_server.RmsBotServerHandler {
	return &service{
		server: server,
		l:      logger.DefaultLogger.Fields(map[string]interface{}{"from": "service"}),
	}
}
