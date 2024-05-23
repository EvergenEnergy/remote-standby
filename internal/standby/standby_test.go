package standby_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/EvergenEnergy/remote-standby/internal/config"
	internalMQTT "github.com/EvergenEnergy/remote-standby/internal/mqtt"
	"github.com/EvergenEnergy/remote-standby/internal/standby"
	"github.com/EvergenEnergy/remote-standby/internal/storage"
	"github.com/stretchr/testify/assert"
)

var (
	testLogger     = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	defaultTimeout = 10 * time.Second
)

func getTestConfig() config.Config {
	return config.Config{
		MQTT: config.MQTTConfig{
			BrokerURL:        "tcp://localhost:1883",
			StandbyTopic:     "cmd/site/standby/serial/#",
			ReadCommandTopic: "cmd/site/handler/serial/cloud",
		},
		Standby: config.StandbyConfig{
			CheckInterval:   time.Duration(1 * time.Second),
			OutageThreshold: time.Duration(2 * time.Second),
		},
	}
}

func TestSubscribesToTopics(t *testing.T) {
	cfg := getTestConfig()
	mqttClient := internalMQTT.MockClient{}
	storageSvc := storage.NewService(testLogger)
	svc := standby.NewService(testLogger, cfg, storageSvc, &mqttClient)

	t.Log("about to run runmqtt")
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	err := svc.Start(ctx)
	t.Log("started runmqtt")
	assert.NoError(t, err)
	time.Sleep(time.Second)
	assert.EqualValues(t, mqttClient.SubscribedTopics, []string{cfg.MQTT.StandbyTopic, cfg.MQTT.ReadCommandTopic})
	svc.Stop()
}

func TestRunsAClient_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}
	cfg := getTestConfig()
	mqttClient := internalMQTT.NewClient(cfg)
	storageSvc := storage.NewService(testLogger)
	svc := standby.NewService(testLogger, cfg, storageSvc, mqttClient)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	err := svc.Start(ctx)
	assert.NoError(t, err)
	svc.Stop()
}

func TestDetectsOutage_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	cfg := getTestConfig()
	mqttClient := internalMQTT.NewClient(cfg)
	storageSvc := storage.NewService(testLogger)
	svc := standby.NewService(testLogger, cfg, storageSvc, mqttClient)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	err := svc.Start(ctx)
	assert.NoError(t, err)
	time.Sleep(time.Second)

	// First check after 1 second, remain in standby
	svc.CheckForOutage(time.Now())
	assert.True(t, svc.InStandbyMode())

	// After 2 seconds, command timestamp exceeds threshold, switch to command mode
	time.Sleep(time.Second)
	svc.CheckForOutage(time.Now())
	assert.True(t, svc.InCommandMode())

	// After 3 seconds, command timestamp still exceeds threshold, remain in command mode
	time.Sleep(time.Second)
	svc.CheckForOutage(time.Now())
	assert.True(t, svc.InCommandMode())

	encPayload, _ := json.Marshal(map[string]interface{}{"action": "test", "value": 23})
	token := mqttClient.Publish(cfg.MQTT.ReadCommandTopic, 1, false, encPayload)
	token.Wait()
	time.Sleep(time.Second)
	testLogger.Info("published to", "topic", cfg.MQTT.ReadCommandTopic)

	// After 4 seconds, new command received, resume standby mode
	svc.CheckForOutage(time.Now())
	assert.True(t, svc.InStandbyMode())

	svc.Stop()
}
