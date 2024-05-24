package standby_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/EvergenEnergy/remote-standby/internal/config"
	mqtt "github.com/EvergenEnergy/remote-standby/internal/mqtt"
	"github.com/EvergenEnergy/remote-standby/internal/plan"
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
			StandbyTopic:     "cmd/site/standby/serial/plan",
			ReadCommandTopic: "cmd/site/handler/serial/cloud",
		},
		Standby: config.StandbyConfig{
			CheckInterval:   time.Duration(1 * time.Second),
			OutageThreshold: time.Duration(2 * time.Second),
			BackupFile:      "/tmp/backup-plan.json",
		},
	}
}

func TestUpdatesTimestamp_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}
	cfg := getTestConfig()
	mqttClient := mqtt.NewClient(cfg)
	storageSvc := storage.NewService(testLogger)
	svc := standby.NewService(testLogger, cfg, storageSvc, mqttClient)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	err := svc.Start(ctx)
	assert.NoError(t, err)

	timestampBeforeCommand := storageSvc.GetCommandTimestamp()

	encPayload, _ := json.Marshal(map[string]interface{}{"action": "test", "value": 23})
	token := mqttClient.Publish(cfg.MQTT.ReadCommandTopic, 1, false, encPayload)
	token.Wait()
	svc.Stop()
	time.Sleep(time.Second)

	timestampAfterCommand := storageSvc.GetCommandTimestamp()
	assert.True(t, timestampAfterCommand.After(timestampBeforeCommand))
}

var optPlan = plan.OptimisationPlan{
	SiteID:       "test-site",
	SetpointType: 1,
	OptimisationIntervals: []plan.OptimisationInterval{
		{
			Interval: plan.OptimisationIntervalTimestamp{
				StartTime: plan.OptimisationTimestamp{Seconds: 1715319000},
				EndTime:   plan.OptimisationTimestamp{Seconds: 1715319900},
			},
			BatteryPower: plan.OptimisationValue{
				Value: 100,
				Unit:  2,
			},
			StateOfCharge: 0.55,
			MeterPower: plan.OptimisationValue{
				Value: 400,
				Unit:  2,
			},
		},
	},
}

func TestStoresAPlan_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}
	cfg := getTestConfig()
	mqttClient := mqtt.NewClient(cfg)
	storageSvc := storage.NewService(testLogger)
	svc := standby.NewService(testLogger, cfg, storageSvc, mqttClient)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	err := svc.Start(ctx)
	assert.NoError(t, err)

	encPayload, _ := json.Marshal(optPlan)
	token := mqttClient.Publish(cfg.MQTT.StandbyTopic, 1, false, encPayload)
	token.Wait()
	time.Sleep(time.Second)

	// TODO once plan replay is implemented, add asserts to indicate we've read the plan
	svc.Stop()
}

func TestDetectsOutage_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	cfg := getTestConfig()
	mqttClient := mqtt.NewClient(cfg)
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
