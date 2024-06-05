package config_test

import (
	"testing"

	"github.com/EvergenEnergy/remote-standby/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestConfig() config.Config {
	return config.Config{
		SiteName:     "test",
		SerialNumber: "device",
		MQTT: config.MQTTConfig{
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

func TestReadFromFile_WhenConfigFilePathExists_ReadsConfig(t *testing.T) {
	testPath := "../../tests/integration/config.yaml"
	require.FileExists(t, testPath)

	cfgGoodPath := config.Config{
		SiteName:          "test",
		SerialNumber:      "device",
		ConfigurationPath: testPath,
	}

	got, err := cfgGoodPath.NewFromFile()
	require.NoError(t, err)

	assert.NotEmpty(t, got)
	assert.Equal(t, got.MQTT.CommandAction, "STORAGE_POINT")
	assert.Equal(t, got.Standby.BackupFile, "/command-standby/backup/plan.json")
}

func TestReadFromFile_WhenConfigFilePathIsNonexistent_ReturnsError(t *testing.T) {
	testPath := "no/such/file"
	require.NoFileExists(t, testPath)

	cfgBadPath := config.Config{
		SiteName:          "test",
		SerialNumber:      "device",
		ConfigurationPath: testPath,
	}

	_, err := cfgBadPath.NewFromFile()
	assert.Error(t, err)
}

func TestReadFromFile_WhenConfigFilePathIsAbsent_ReturnsError(t *testing.T) {
	cfgNoPath := config.Config{
		SiteName:     "test",
		SerialNumber: "device",
	}

	_, err := cfgNoPath.NewFromFile()
	assert.Error(t, err)
}
