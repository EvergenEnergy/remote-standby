package standby

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/EvergenEnergy/remote-standby/internal/config"
	"github.com/stretchr/testify/assert"
)

var testLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

func getTestConfig() config.Config {
	return config.Config{
		MQTT: config.MQTTConfig{
			BrokerURL:    "tcp://localhost:1883",
			StandbyTopic: "cmd/site/standby/serial/#",
		},
		Standby: config.StandbyConfig{
			CheckInterval:   time.Duration(2 * time.Second),
			OutageThreshold: time.Duration(4 * time.Second),
		},
	}
}

func TestInitsAClient(t *testing.T) {
	cfg := getTestConfig()
	svc := NewService(testLogger, cfg)
	assert.NotEmpty(t, svc.mqttClient)
}

func TestRunsAClient_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}
	cfg := getTestConfig()

	svc := NewService(testLogger, cfg)
	err := svc.RunMQTT(context.Background())
	assert.NoError(t, err)
	svc.StopMQTT()
}

func TestDetectsOutage(t *testing.T) {
	cfg := getTestConfig()
	svc := NewService(testLogger, cfg)

	// Start at current time, latest command received 1 second ago
	interval := time.Duration(2 * time.Second)
	currentTime := time.Now()
	svc.latestCommandReceived = currentTime.Add(-1 * time.Second)

	// First check after 1 interval, command is recent, remain in standby
	currentTime = currentTime.Add(1 * interval)
	svc.CheckForOutage(currentTime)
	assert.EqualValues(t, svc.mode, StandbyMode)

	// After 2 intervals, command timestamp exceeds threshold, switch to command mode
	currentTime = currentTime.Add(2 * interval)
	svc.CheckForOutage(currentTime)
	assert.EqualValues(t, svc.mode, CommandMode)

	// After 3 intervals, command timestamp still exceeds threshold, remain in command mode
	currentTime = currentTime.Add(3 * interval)
	svc.CheckForOutage(currentTime)
	assert.EqualValues(t, svc.mode, CommandMode)

	// After 4 intervals, new command received, resume standby mode
	currentTime = currentTime.Add(4 * interval)
	svc.latestCommandReceived = currentTime.Add(-1 * time.Second)
	svc.CheckForOutage(currentTime)
	assert.EqualValues(t, svc.mode, StandbyMode)
}
