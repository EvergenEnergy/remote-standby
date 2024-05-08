package config_test

import (
	"testing"

	"github.com/EvergenEnergy/remote-standby/config"

	"github.com/stretchr/testify/assert"
)

func getTestConfig(t *testing.T) config.Config {
	t.Setenv("MQTT_BROKER_URL", "tcp://localhost:1883")
	t.Setenv("MQTT_STANDBY_TOPIC", "cmd/site/standby/serial/#")

	cfg, err := config.FromEnv()
	assert.NoError(t, err)
	return cfg
}

func TestReadConfig(t *testing.T) {
	cfg := getTestConfig(t)
	assert.NotEmpty(t, cfg)
}
