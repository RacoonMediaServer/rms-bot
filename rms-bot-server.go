package main

import (
	"fmt"
	"github.com/RacoonMediaServer/rms-bot-server/internal/bot"
	"github.com/RacoonMediaServer/rms-bot-server/internal/config"
	"github.com/RacoonMediaServer/rms-bot-server/internal/db"
	"github.com/RacoonMediaServer/rms-bot-server/internal/server"
	botService "github.com/RacoonMediaServer/rms-bot-server/internal/service"
	rms_bot_server "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-bot-server"
	"github.com/RacoonMediaServer/rms-packages/pkg/service/servicemgr"
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
)

var Version = "v0.0.0"

const serviceName = "rms-bot-server"

func main() {
	logger.Infof("%s %s", serviceName, Version)
	defer logger.Info("DONE.")

	useDebug := false

	service := micro.NewService(
		micro.Name(serviceName),
		micro.Version(Version),
		micro.Flags(
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"debug"},
				Usage:       "debug log level",
				Value:       false,
				Destination: &useDebug,
			},
		),
	)

	service.Init(
		micro.Action(func(context *cli.Context) error {
			configFile := fmt.Sprintf("/etc/rms/%s.json", serviceName)
			if context.IsSet("config") {
				configFile = context.String("config")
			}
			return config.Load(configFile)
		}),
	)

	if useDebug {
		_ = logger.Init(logger.WithLevel(logger.DebugLevel))
	}

	cfg := config.Config()

	database, err := db.Connect(cfg.Database)
	if err != nil {
		logger.Fatalf("Connect to database failed: %s", err)
	}

	srv := server.New(servicemgr.NewServiceFactory(service))

	if err = rms_bot_server.RegisterRmsBotServerHandler(service.Server(), botService.New(srv, database)); err != nil {
		logger.Fatalf("Register service failed: %s", err)
	}

	// запускаем Telegram бот
	tBot, err := bot.NewBot(cfg.Bot.Token, database, srv)
	if err != nil {
		logger.Fatalf("Cannot start Telegram bot: %s", err)
	}
	defer tBot.Stop()

	// запускам сервер, который будет обрабатывать WebSocket подключения от клиентов
	if err = srv.ListenAndServe(cfg.Http.Host, cfg.Http.Port); err != nil {
		logger.Fatalf("Cannot start server: %s", err)
	}
	srv.Wait()
}
