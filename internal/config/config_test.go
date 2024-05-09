package config_test

import (
	"testing"

	"github.com/EvergenEnergy/remote-standby/internal/config"

	"github.com/stretchr/testify/assert"
)

/*
func getTestConfigFromEnv(t *testing.T) config.Config {
	t.Setenv("MQTT_BROKER_URL", "tcp://localhost:1883")
	t.Setenv("MQTT_STANDBY_TOPIC", "cmd/site/standby/serial/#")

	cfg, err := config.FromEnv()
	assert.NoError(t, err)
	return cfg
}
*/

func getTestConfig() config.Config {
	return config.Config{
		MQTT: config.MQTTConfig{
			BrokerURL:    "tcp://localhost:1833",
			StandbyTopic: "cmd/site/standby/serial/#",
		},
	}
}

func TestReadConfig(t *testing.T) {
	cfg := getTestConfig()
	assert.NotEmpty(t, cfg)
}
