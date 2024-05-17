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

func TestChecksForOutage(t *testing.T) {
	cfg := getTestConfig()
	svc := NewService(testLogger, cfg)
	svc.latestCommandReceived = time.Now()
	assert.EqualValues(t, svc.mode, StandbyMode)
	time.Sleep(4 * time.Second)
	svc.CheckForOutage()
	assert.EqualValues(t, svc.mode, CommandMode)
}
