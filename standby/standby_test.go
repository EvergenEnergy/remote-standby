package standby_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/EvergenEnergy/remote-standby/config"
	"github.com/EvergenEnergy/remote-standby/standby"
	"github.com/stretchr/testify/assert"
)

var logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

func getConfig(t *testing.T) config.Config {
	t.Setenv("MQTT_BROKER_URL", "tcp://localhost:1883")
	t.Setenv("MQTT_STANDBY_TOPIC", "cmd/site/standby/serial/#")

	cfg, err := config.FromEnv()
	assert.NoError(t, err)
	return cfg
}

func TestInitsAClient(t *testing.T) {
	cfg := getConfig(t)
	svc := standby.NewService(logger, cfg)
	assert.NotEmpty(t, svc.MQTTClient)
}

func TestRunsAClient(t *testing.T) {
	cfg := getConfig(t)

	svc := standby.NewService(logger, cfg)
	err := svc.RunMQTT(context.Background())
	assert.NoError(t, err)
	svc.StopMQTT()
}
