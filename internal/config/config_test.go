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

func TestReadFromFile(t *testing.T) {
	cfgNoPath := config.Config{
		SiteName:     "test",
		SerialNumber: "device",
	}
	_, err := cfgNoPath.NewFromFile()
	assert.Error(t, err)

	cfgBadPath := config.Config{
		SiteName:          "test",
		SerialNumber:      "device",
		ConfigurationPath: "no/such/file",
	}
	_, err = cfgBadPath.NewFromFile()
	assert.Error(t, err)

	cfgGoodPath := config.Config{
		SiteName:          "test",
		SerialNumber:      "device",
		ConfigurationPath: "../../tests/integration/config.yaml",
	}
	t.Log(config.DumpYAML("../../tests/integration/config.yaml"))
	fileCfg, err := cfgGoodPath.NewFromFile()
	t.Log("file config has site name ", fileCfg.SiteName)

	assert.NotEmpty(t, fileCfg)
	assert.NoError(t, err)
	assert.True(t, false)

	assert.Contains(t, fileCfg.MQTT.CommandAction, "STORAGE_POINT")
	assert.Contains(t, fileCfg.Standby.BackupFile, "/command-standby/backup/plan.json")
}
