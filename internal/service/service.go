package service

import (
	"context"
	rms_bot_server "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-bot-server"
	"google.golang.org/protobuf/types/known/emptypb"
)

type service struct {
}

func (s service) DropSession(ctx context.Context, request *rms_bot_server.DropSessionRequest, empty *emptypb.Empty) error {
	//TODO implement me
	panic("implement me")
}

func New() rms_bot_server.RmsBotServerHandler {
	return &service{}
}
