package config

import (
	"github.com/RacoonMediaServer/rms-bot-server/internal/comm"
	"github.com/RacoonMediaServer/rms-packages/pkg/configuration"
)

// Bot is settings of Telegram Bot
type Bot struct {
	Token string
}

// Configuration represents entire service configuration
type Configuration struct {
	Database configuration.Database
	Http     configuration.Http
	Bots     map[string]*Bot
	Debug    configuration.Debug
}

var config Configuration

// Load open and parses configuration file
func Load(configFilePath string) error {
	return configuration.Load(configFilePath, &config)
}

// Config returns loaded configuration
func Config() Configuration {
	return config
}

func (c Configuration) Endpoints() []comm.Endpoint {
	result := make([]comm.Endpoint, 0, len(c.Bots))
	for k := range c.Bots {
		e := comm.Endpoint{ID: k}
		result = append(result, e)
	}
	return result
}
