package config_test

import (
	"testing"

	"github.com/EvergenEnergy/remote-standby/internal/config"

	"github.com/stretchr/testify/assert"
)

func getTestConfig() config.Config {
	return config.Config{
		SiteName:     "test",
		SerialNumber: "device",
		MQTT: config.MQTTConfig{
			BrokerURL:    "tcp://localhost:1833",
			StandbyTopic: "cmd/${SITE_NAME}/standby/${SERIAL_NUMBER}/#",
		},
	}
}

func TestInterpolateVars(t *testing.T) {
	cfg := getTestConfig()
	assert.NotEmpty(t, cfg)

	assert.Contains(t, cfg.MQTT.StandbyTopic, "${SITE_NAME}")
	assert.Contains(t, cfg.MQTT.StandbyTopic, "${SERIAL_NUMBER}")

	cfg.InterpolateEnvVars()
	assert.NotContains(t, cfg.MQTT.StandbyTopic, "${SITE_NAME}")
	assert.NotContains(t, cfg.MQTT.StandbyTopic, "${SERIAL_NUMBER}")
	assert.EqualValues(t, cfg.MQTT.StandbyTopic, "cmd/test/standby/device/#")
}
