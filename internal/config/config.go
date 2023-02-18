package config

import "github.com/RacoonMediaServer/rms-packages/pkg/configuration"

// Bot is settings of Telegram Bot
type Bot struct {
	Token string
}

// Configuration represents entire service configuration
type Configuration struct {
	Database configuration.Database
	Http     configuration.Http
	Monitor  configuration.Monitor
	Bot      Bot
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
