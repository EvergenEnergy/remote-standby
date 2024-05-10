package config

import (
	"fmt"
	"strings"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
)

// Config holds all configurable values of the service.
// It can contain values specified either in the environment or by file,
// but file-based values cannot be required.
type Config struct {
	SiteName          string        `env:"SITE_NAME" required:"true"`
	SerialNumber      string        `env:"SERIAL_NUMBER" required:"true"`
	ConfigurationPath string        `env:"CONFIGURATION_PATH" default:"config.yaml"`
	Logging           LoggingConfig `yaml:"logging"`
	MQTT              MQTTConfig    `yaml:"mqtt"`
}

type LoggingConfig struct {
	Level string `yaml:"level" default:"info"`
}

type MQTTConfig struct {
	BrokerURL    string `yaml:"broker_url" default:"tcp://localhost:1883/"`
	CommandTopic string `yaml:"command_topic" default:"cmd/${SITE_NAME}/handler/${SERIAL_NUMBER}/#"`
	StandbyTopic string `yaml:"standby_topic" default:"cmd/${SITE_NAME}/standby/${SERIAL_NUMBER}/#"`
}

func fromEnv() (Config, error) {
	var cfg Config

	if err := aconfig.LoaderFor(&cfg, aconfig.Config{}).Load(); err != nil {
		return Config{}, fmt.Errorf("unable to parse config: %w", err)
	}

	return cfg, nil
}

func FromFile() (Config, error) {
	var cfg Config

	// read config from env vars first
	configEnv, err := fromEnv()
	if err != nil {
		return Config{}, fmt.Errorf("unable to read config from env: %w", err)
	}

	// now read from both env and file, using the
	// config path specified in the env var
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		SkipFlags: true,
		Files:     []string{configEnv.ConfigurationPath},
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
