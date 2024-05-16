package standby_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/EvergenEnergy/remote-standby/internal/config"
	"github.com/EvergenEnergy/remote-standby/internal/standby"
	"github.com/stretchr/testify/assert"
)

var logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

func getTestConfig() config.Config {
	return config.Config{
		MQTT: config.MQTTConfig{
			BrokerURL:    "tcp://mosquitto:1883",
			StandbyTopic: "cmd/site/standby/serial/#",
		},
		Standby: config.StandbyConfig{
			CheckInterval:  2,
			OutageInterval: 4,
		},
	}
}

func TestInitsAClient(t *testing.T) {
	cfg := getTestConfig()
	svc := standby.NewService(logger, cfg)
	assert.NotEmpty(t, svc.MQTTClient)
}

func TestRunsAClient_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}
	cfg := getTestConfig()

	svc := standby.NewService(logger, cfg)
	err := svc.RunMQTT(context.Background())
	assert.NoError(t, err)
	svc.StopMQTT()
}

func TestChecksForOutage(t *testing.T) {
	cfg := getTestConfig()
	svc := standby.NewService(logger, cfg)
	svc.LatestCommandReceived = time.Now()
	assert.EqualValues(t, svc.Mode, standby.StandbyMode)
	time.Sleep(4 * time.Second)
	svc.CheckForOutage(time.Duration(cfg.Standby.OutageInterval) * time.Second)
	assert.EqualValues(t, svc.Mode, standby.CommandMode)
}
