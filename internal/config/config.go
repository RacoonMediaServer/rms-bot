package config

import "github.com/RacoonMediaServer/rms-packages/pkg/configuration"

// Bot is settings of Telegram Bot
type Bot struct {
	Token            string
	SelfRegistration bool `json:"selfRegistration"`
}

// Configuration represents entire service configuration
type Configuration struct {
	Database configuration.Database
	Http     configuration.Http
	Bot      Bot
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
