package config

import (
	"fmt"
	"strings"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
)

// Config holds all configurable values of the service.
type Config struct {
	SiteName     string        `env:"SITE_NAME" required:"true"`
	SerialNumber string        `env:"SERIAL_NUMBER" required:"true"`
	Logging      LoggingConfig `yaml:"Logging"`
	MQTT         MQTTConfig    `yaml:"MQTT"`
}

// LoggingConfig is a config for logger.
type LoggingConfig struct {
	Level string `yaml:"Level" default:"info"`
}

// MQTTConfig is a config for logger.
type MQTTConfig struct {
	BrokerURL    string `yaml:"BrokerUrl"`
	CommandTopic string `yaml:"CommandTopic"`
	StandbyTopic string `yaml:"StandbyTopic"`
}

// FromEnv creates new config based on environment variables.
func FromEnv() (Config, error) {
	var cfg Config

	if err := aconfig.LoaderFor(&cfg, aconfig.Config{}).Load(); err != nil {
		return Config{}, fmt.Errorf("unable to parse config: %w", err)
	}

	return cfg, nil
}

func FromFile() (Config, error) {
	var cfg Config

	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		SkipFlags: true,
		Files:     []string{"config/config.yaml"},
		FileDecoders: map[string]aconfig.FileDecoder{
			".yaml": aconfigyaml.New(),
		},
	})
	if err := loader.Load(); err != nil {
		return Config{}, fmt.Errorf("unable to parse config: %w", err)
	}

	cfg.interpolateEnvVars()

	return cfg, nil
}

func (cfg *Config) interpolateEnvVars() {
	cfg.MQTT.CommandTopic = strings.Replace(cfg.MQTT.CommandTopic, "${SITE_NAME}", cfg.SiteName, -1)
	cfg.MQTT.CommandTopic = strings.Replace(cfg.MQTT.CommandTopic, "${SERIAL_NUMBER}", cfg.SerialNumber, -1)
	cfg.MQTT.StandbyTopic = strings.Replace(cfg.MQTT.StandbyTopic, "${SITE_NAME}", cfg.SiteName, -1)
	cfg.MQTT.StandbyTopic = strings.Replace(cfg.MQTT.StandbyTopic, "${SERIAL_NUMBER}", cfg.SerialNumber, -1)
}
