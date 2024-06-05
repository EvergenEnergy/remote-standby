package outagelog_test

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/EvergenEnergy/remote-standby/internal/config"
	"github.com/EvergenEnergy/remote-standby/internal/mqtt"
	"github.com/EvergenEnergy/remote-standby/internal/outagelog"
	"github.com/EvergenEnergy/remote-standby/internal/publisher"
	"github.com/EvergenEnergy/remote-standby/internal/standby"
	"github.com/EvergenEnergy/remote-standby/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testLogger     = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	defaultTimeout = 10 * time.Second
)

func getTestConfig() config.Config {
	tmpLogPath := fmt.Sprintf("/tmp/outage-%d.log", time.Now().Nanosecond())
	return config.Config{
		SiteName:     "test",
		SerialNumber: "device",
		MQTT: config.MQTTConfig{
			BrokerURL:         "tcp://localhost:1883",
			ReadCommandTopic:  "cmd/site/handler/serial/cloud",
			WriteCommandTopic: "cmd/site/handler/serial/standby",
			StandbyTopic:      "cmd/site/standby/serial/plan",
		},
		Standby: config.StandbyConfig{
			CheckInterval:   time.Duration(1 * time.Second),
			OutageThreshold: time.Duration(2 * time.Second),
			OutageLogFile:   tmpLogPath,
		},
	}
}

func TestWriteOutageLog(t *testing.T) {
	cfg := getTestConfig()
	logPath := cfg.Standby.OutageLogFile
	defer os.Remove(logPath)

	logHandle, err := outagelog.Open(logPath)
	require.NoError(t, err)
	logHandler := outagelog.NewHandler(logHandle, testLogger)

	logHandler.Append("test message", nil)
	logHandler.Append("test message with details", map[string]string{"foo": "baa", "num": "23"})
	logHandler.Close()

	logLines := readLogFile(t, logPath)

	assert.Len(t, logLines, 2)
	assert.Contains(t, logLines[0], "test message")
	assert.Contains(t, logLines[1], "foo=baa")
	assert.Contains(t, logLines[1], "num=23")
}

func readLogFile(t *testing.T, logPath string) []string {
	content, err := os.Open(logPath)
	require.NoError(t, err)

	scanner := bufio.NewScanner(content)

	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}

func TestWriteLogDuringOutage_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}
	cfg := getTestConfig()
	logPath := cfg.Standby.OutageLogFile
	defer os.Remove(logPath)

	mqttClient := mqtt.NewClient(cfg)
	storageSvc := storage.NewService(testLogger)
	publisherSvc := publisher.NewService(testLogger, cfg, mqttClient)
	logHandle, err := outagelog.Open(logPath)
	require.NoError(t, err)
	logHandler := outagelog.NewHandler(logHandle, testLogger)
	assert.NoError(t, err)
	svc := standby.NewService(testLogger, cfg, storageSvc, publisherSvc, logHandler, mqttClient)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	err = svc.Start(ctx)
	assert.NoError(t, err)
	// sleep for long enough for an outage to be triggered (including some buffer)
	time.Sleep(cfg.Standby.CheckInterval + cfg.Standby.OutageThreshold + 1*time.Second)
	svc.Stop()

	logLines := readLogFile(t, logPath)
	assert.GreaterOrEqual(t, len(logLines), 4)
	assert.Contains(t, logLines[0], "Service started")
	assert.Contains(t, logLines[1], "Entered command mode")
	assert.Contains(t, logLines[2], "No command available")

	assert.NotEmpty(t, logLines)
}
