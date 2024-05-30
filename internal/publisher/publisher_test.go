package publisher_test

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/EvergenEnergy/remote-standby/internal/config"
	"github.com/EvergenEnergy/remote-standby/internal/mqtt"
	"github.com/EvergenEnergy/remote-standby/internal/plan"
	"github.com/EvergenEnergy/remote-standby/internal/publisher"
	"github.com/stretchr/testify/assert"
)

var testLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

func TestBuildCommandPayloads(t *testing.T) {
	type test struct {
		meterPower float32
		meterUnit  int
		expected   float64
	}

	tests := []test{
		{meterPower: 1234, meterUnit: 1, expected: 1.234},
		{meterPower: 1234, meterUnit: 2, expected: 1234},
		{meterPower: 1.234, meterUnit: 3, expected: 1234},
	}

	for _, tc := range tests {
		payload := publisher.BuildCommandPayload("actionvalue", plan.OptimisationInterval{
			MeterPower: plan.OptimisationValue{Value: tc.meterPower, Unit: tc.meterUnit},
		})
		assert.InDelta(t, tc.expected, payload.Value, 0.0001)
	}
}

// The PublishCommand and PublishError methods are more thoroughly tested in the standby_test package.

func getTestConfig() config.Config {
	return config.Config{
		MQTT: config.MQTTConfig{
			BrokerURL:         "tcp://localhost:1883",
			StandbyTopic:      "cmd/site/standby/serial/plan",
			ErrorTopic:        "cmd/site/error/serial/error",
			ReadCommandTopic:  "cmd/site/handler/serial/cloud",
			WriteCommandTopic: "cmd/site/handler/serial/standby",
			CommandAction:     "STORAGEPOINT",
		},
		Standby: config.StandbyConfig{
			CheckInterval:   time.Duration(1 * time.Second),
			OutageThreshold: time.Duration(2 * time.Second),
			BackupFile:      "/tmp/backup-plan.json",
		},
	}
}

func TestPublishesCommand(t *testing.T) {
	cfg := getTestConfig()
	mqttClient := mqtt.NewClient(cfg)
	publisherSvc := publisher.NewService(testLogger, cfg, mqttClient)

	err := publisherSvc.PublishCommand(plan.OptimisationInterval{
		Interval: plan.OptimisationIntervalTimestamp{
			StartTime: plan.OptimisationTimestamp{
				Seconds: time.Now().Unix(),
			},
			EndTime: plan.OptimisationTimestamp{
				Seconds: time.Now().Add(10 * time.Second).Unix(),
			},
		},
		BatteryPower: plan.OptimisationValue{
			Value: 100,
			Unit:  2,
		},
		StateOfCharge: 0.8,
		MeterPower: plan.OptimisationValue{
			Value: 50,
			Unit:  2,
		},
	},
	)
	assert.NoError(t, err)
}

func TestPublishesError(t *testing.T) {
	cfg := getTestConfig()
	cfg.MQTT.CommandAction = ""

	mqttClient := mqtt.NewClient(cfg)
	publisherSvc := publisher.NewService(testLogger, cfg, mqttClient)

	err := publisherSvc.PublishCommand(plan.OptimisationInterval{})
	publisherSvc.PublishError("something went wrong publishing a command", err)
	assert.Error(t, err)
}
