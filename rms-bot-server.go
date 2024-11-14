package main

import (
	"fmt"
	"net/http"

	"github.com/RacoonMediaServer/rms-bot-server/internal/bot"
	"github.com/RacoonMediaServer/rms-bot-server/internal/config"
	"github.com/RacoonMediaServer/rms-bot-server/internal/db"
	"github.com/RacoonMediaServer/rms-bot-server/internal/server"
	botService "github.com/RacoonMediaServer/rms-bot-server/internal/service"
	rms_bot_server "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-bot-server"
	"github.com/RacoonMediaServer/rms-packages/pkg/service/servicemgr"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"

	// Plugins
	_ "github.com/go-micro/plugins/v4/registry/etcd"
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

	cfg := config.Config()

	if useDebug || cfg.Debug.Verbose {
		_ = logger.Init(logger.WithLevel(logger.DebugLevel))
	}

	endpoints := cfg.Endpoints()
	if len(endpoints) == 0 {
		logger.Fatal("No endpoints specified")
	}

	database, err := db.Connect(cfg.Database)
	if err != nil {
		logger.Fatalf("Connect to database failed: %s", err)
	}

	srv := server.New(servicemgr.NewServiceFactory(service), endpoints)

	if err = rms_bot_server.RegisterRmsBotServerHandler(service.Server(), botService.New(srv, database)); err != nil {
		logger.Fatalf("Register service failed: %s", err)
	}

	// запускаем Telegram бот
	for id, botConfig := range cfg.Bots {
		endpoint, err := srv.GetEndpoint(id)
		if err != nil {
			logger.Fatalf("Endpoint wasn't registered properly: %s", err)
		}
		tBot, err := bot.NewBot(botConfig.Token, database, endpoint)
		if err != nil {
			logger.Fatalf("Cannot start Telegram bot: %s", err)
		}
		defer tBot.Stop()
	}

	// запускам сервер, который будет обрабатывать WebSocket подключения от клиентов
	if err = srv.ListenAndServe(cfg.Http.Host, cfg.Http.Port); err != nil {
		logger.Fatalf("Cannot start server: %s", err)
	}

	monitorConfig := cfg.Debug.Monitor
	if monitorConfig.Enabled {
		go func() {
			http.Handle("/metrics", promhttp.Handler())
			if err := http.ListenAndServe(fmt.Sprintf("%s:%d", monitorConfig.Host, monitorConfig.Port), nil); err != nil {
				logger.Fatalf("Cannot bind monitoring endpoint: %s", err)
			}
		}()
	}

	srv.Wait()
}
