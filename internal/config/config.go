package config

import (
	"fmt"
	"strings"
	"time"

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
	Standby           StandbyConfig `yaml:"standby"`
}

type LoggingConfig struct {
	Level string `yaml:"level" default:"info"`
}

type MQTTConfig struct {
	BrokerURL         string `yaml:"broker_url" default:"tcp://localhost:1883/"`
	WriteCommandTopic string `yaml:"write_command_topic" default:"cmd/${SITE_NAME}/handler/${SERIAL_NUMBER}/standby"`
	ReadCommandTopic  string `yaml:"read_command_topic" default:"cmd/${SITE_NAME}/handler/${SERIAL_NUMBER}/cloud"`
	StandbyTopic      string `yaml:"standby_topic" default:"cmd/${SITE_NAME}/standby/${SERIAL_NUMBER}/#"`
	ErrorTopic        string `yaml:"error_topic" default:"dt/${SITE_NAME}/error/${SERIAL_NUMBER}"`
}

type StandbyConfig struct {
	BackupFile      string        `yaml:"backup_file" default:"plan.json"`
	CheckInterval   time.Duration `yaml:"check_interval" default:"60s"`
	OutageThreshold time.Duration `yaml:"outage_interval" default:"180s"`
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

	cfg.InterpolateEnvVars()

	return cfg, nil
}

func (cfg *Config) InterpolateEnvVars() {
	replacer := strings.NewReplacer("${SITE_NAME}", cfg.SiteName, "${SERIAL_NUMBER}", cfg.SerialNumber)
	cfg.MQTT.ReadCommandTopic = replacer.Replace(cfg.MQTT.ReadCommandTopic)
	cfg.MQTT.WriteCommandTopic = replacer.Replace(cfg.MQTT.WriteCommandTopic)
	cfg.MQTT.StandbyTopic = replacer.Replace(cfg.MQTT.StandbyTopic)
	cfg.MQTT.ErrorTopic = replacer.Replace(cfg.MQTT.ErrorTopic)
}
