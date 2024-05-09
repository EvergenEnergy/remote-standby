package standby_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/EvergenEnergy/remote-standby/internal/config"
	"github.com/EvergenEnergy/remote-standby/internal/standby"
	"github.com/stretchr/testify/assert"
)

var logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

func getTestConfig() config.Config {
	return config.Config{
		MQTT: config.MQTTConfig{
			BrokerURL:    "tcp://localhost:1833",
			StandbyTopic: "cmd/site/standby/serial/#",
		},
	}
}

func TestInitsAClient(t *testing.T) {
	cfg := getTestConfig()
	svc := standby.NewService(logger, cfg)
	assert.NotEmpty(t, svc.MQTTClient)
}

/*
func TestRunsAClient(t *testing.T) {
	cfg := getTestConfig()

	svc := standby.NewService(logger, cfg)
	err := svc.RunMQTT(context.Background())
	assert.NoError(t, err)
	svc.StopMQTT()
}
*/
