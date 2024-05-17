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
	currentTime := time.Now()

	threeSecondsAgo := time.Duration(-3 * time.Second)
	svc.latestCommandReceived = currentTime.Add(threeSecondsAgo)
	svc.CheckForOutage(currentTime)
	assert.EqualValues(t, svc.mode, StandbyMode)

	fiveSecondsAgo := time.Duration(-5 * time.Second)
	svc.latestCommandReceived = currentTime.Add(fiveSecondsAgo)
	svc.CheckForOutage(currentTime)
	assert.EqualValues(t, svc.mode, CommandMode)
}
