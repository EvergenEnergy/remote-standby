package config

import (
	"fmt"

	"github.com/cristalhq/aconfig"
)

// Config holds all configurable values of the service.
type Config struct {
	SiteName   string        `env:"SITE_NAME" required:"true"`
	DeviceName string        `env:"DEVICE_NAME" required:"true"`
	Logging    LoggingConfig `env:"LOGGING"`
	MQTT       MQTTConfig    `env:"MQTT"`
}

// LoggingConfig is a config for logger.
type LoggingConfig struct {
	Level string `env:"LEVEL" default:"info"`
}

// MQTTConfig is a config for logger.
type MQTTConfig struct {
	BrokerURL    string `env:"BROKER_URL"`
	CommandTopic string `env:"COMMAND_TOPIC"`
	StandbyTopic string `env:"STANDBY_TOPIC"`
}

// FromEnv creates new config based on environment variables.
func FromEnv() (Config, error) {
	var cfg Config

	if err := aconfig.LoaderFor(&cfg, aconfig.Config{}).Load(); err != nil {
		return Config{}, fmt.Errorf("unable to parse config: %w", err)
	}

	return cfg, nil
}
