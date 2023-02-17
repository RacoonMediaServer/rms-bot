package config

import "github.com/RacoonMediaServer/rms-packages/pkg/configuration"

// HTTP is settings of web server
type HTTP struct {
	Host string
	Port int
}

// Bot is settings of Telegram Bot
type Bot struct {
	Token string
}

// Configuration represents entire service configuration
type Configuration struct {
	Database configuration.Database
	Http     HTTP
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
