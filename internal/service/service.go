package service

import (
	"context"
	rms_bot_server "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-bot-server"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server interface {
	DropSession(token string)
}

type service struct {
	server Server
}

func (s service) DropSession(ctx context.Context, request *rms_bot_server.DropSessionRequest, empty *emptypb.Empty) error {
	s.server.DropSession(request.Token)
	return nil
}

func New(server Server) rms_bot_server.RmsBotServerHandler {
	return &service{server: server}
}
