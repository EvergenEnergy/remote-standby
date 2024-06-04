package standby_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/EvergenEnergy/remote-standby/internal/config"
	"github.com/EvergenEnergy/remote-standby/internal/mqtt"
	"github.com/EvergenEnergy/remote-standby/internal/outagelog"
	"github.com/EvergenEnergy/remote-standby/internal/plan"
	"github.com/EvergenEnergy/remote-standby/internal/publisher"
	"github.com/EvergenEnergy/remote-standby/internal/standby"
	"github.com/EvergenEnergy/remote-standby/internal/storage"
	pahoMQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
)

var (
	testLogger     = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	defaultTimeout = 10 * time.Second
)

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
			OutageLogFile:   "/tmp/outage.log",
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
	publisherSvc := publisher.NewService(testLogger, cfg, mqttClient)
	logHandler, err := outagelog.NewHandler(cfg.Standby.OutageLogFile, testLogger)
	assert.NoError(t, err)
	defer logHandler.Close()
	defer os.Remove(cfg.Standby.OutageLogFile)
	svc := standby.NewService(testLogger, cfg, storageSvc, publisherSvc, logHandler, mqttClient)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	err = svc.Start(ctx)
	assert.NoError(t, err)
	defer svc.Stop()

	timestampBeforeCommand := storageSvc.GetCommandTimestamp()

	encPayload, _ := json.Marshal(map[string]interface{}{"action": "test", "value": 23})
	token := mqttClient.Publish(cfg.MQTT.ReadCommandTopic, 1, false, encPayload)
	token.Wait()
	svc.Stop()
	time.Sleep(time.Second)

	timestampAfterCommand := storageSvc.GetCommandTimestamp()
	assert.True(t, timestampAfterCommand.After(timestampBeforeCommand))
}

func TestPublishesError_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	isErrorPublished := false
	mu := new(sync.Mutex)

	getErr := func() bool {
		mu.Lock()
		defer mu.Unlock()
		return isErrorPublished
	}
	setErr := func(errFlag bool) {
		mu.Lock()
		defer mu.Unlock()
		isErrorPublished = errFlag
	}

	cfg := getTestConfig()
	mqttClient := mqtt.NewClient(cfg)
	storageSvc := storage.NewService(testLogger)
	publisherSvc := publisher.NewService(testLogger, cfg, mqttClient)
	logHandler, err := outagelog.NewHandler(cfg.Standby.OutageLogFile, testLogger)
	assert.NoError(t, err)
	defer logHandler.Close()
	defer os.Remove(cfg.Standby.OutageLogFile)
	svc := standby.NewService(testLogger, cfg, storageSvc, publisherSvc, logHandler, mqttClient)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	err = svc.Start(ctx)
	assert.NoError(t, err)
	defer svc.Stop()

	errTopic := fmt.Sprintf("%s/%s", cfg.MQTT.ErrorTopic, "Standby")
	mqttClient.Subscribe(errTopic, 1, func(client pahoMQTT.Client, msg pahoMQTT.Message) {
		setErr(true)
	})

	encPayload, _ := json.Marshal(map[string]interface{}{"not": "an", "optimisation": "plan"})
	mqttClient.Publish(cfg.MQTT.StandbyTopic, 1, false, encPayload)
	time.Sleep(time.Second)

	assert.True(t, getErr())
}

func getOptPlan() plan.OptimisationPlan {
	return plan.OptimisationPlan{
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
}

func TestStoresAndReplaysAPlan_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}
	var commandMsg pahoMQTT.Message
	mu := new(sync.Mutex)

	getMsg := func() pahoMQTT.Message {
		mu.Lock()
		defer mu.Unlock()
		return commandMsg
	}
	setMsg := func(msg pahoMQTT.Message) {
		mu.Lock()
		defer mu.Unlock()
		commandMsg = msg
	}

	cfg := getTestConfig()
	mqttClient := mqtt.NewClient(cfg)
	storageSvc := storage.NewService(testLogger)
	publisherSvc := publisher.NewService(testLogger, cfg, mqttClient)
	logHandler, err := outagelog.NewHandler(cfg.Standby.OutageLogFile, testLogger)
	assert.NoError(t, err)
	defer logHandler.Close()
	defer os.Remove(cfg.Standby.OutageLogFile)
	svc := standby.NewService(testLogger, cfg, storageSvc, publisherSvc, logHandler, mqttClient)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	err = svc.Start(ctx)
	assert.NoError(t, err)
	defer svc.Stop()

	mqttClient.Subscribe(cfg.MQTT.WriteCommandTopic, 1, func(client pahoMQTT.Client, msg pahoMQTT.Message) {
		setMsg(msg)
	})

	// Set the first interval in the optimisation plan to start 10 seconds ago, and publish it
	currentTime := time.Now()
	targetTime := currentTime.Add(-10 * time.Second)
	t.Log("Adjusting interval start time to ", "targetTime", targetTime)
	optPlan := getOptPlan()
	optPlan.OptimisationIntervals[0].Interval.StartTime.Seconds = targetTime.Unix()
	optPlan.OptimisationIntervals[0].Interval.EndTime.Seconds = targetTime.Add(300 * time.Second).Unix()

	encPayload, _ := json.Marshal(optPlan)
	token := mqttClient.Publish(cfg.MQTT.StandbyTopic, 1, false, encPayload)
	token.Wait()

	// wait for the service to detect an outage and send a replacement command
	time.Sleep(4 * time.Second)
	svc.Stop()

	subscribedMsg := getMsg()
	assert.NotEmpty(t, subscribedMsg)

	msg := publisher.CommandPayload{}
	err = json.Unmarshal(subscribedMsg.Payload(), &msg)
	assert.NoError(t, err)

	assert.EqualValues(t, cfg.MQTT.CommandAction, msg.Action)
	assert.EqualValues(t, optPlan.OptimisationIntervals[0].MeterPower.Value, msg.Value)
}

func TestDetectsOutage_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}

	cfg := getTestConfig()
	mqttClient := mqtt.NewClient(cfg)
	storageSvc := storage.NewService(testLogger)
	publisher := publisher.NewService(testLogger, cfg, mqttClient)
	logHandler, err := outagelog.NewHandler(cfg.Standby.OutageLogFile, testLogger)
	assert.NoError(t, err)
	defer logHandler.Close()
	defer os.Remove(cfg.Standby.OutageLogFile)
	svc := standby.NewService(testLogger, cfg, storageSvc, publisher, logHandler, mqttClient)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	err = svc.Start(ctx)
	assert.NoError(t, err)
	defer svc.Stop()

	// First check after 1 second, remain in standby
	time.Sleep(1 * time.Second)
	assert.True(t, svc.InStandbyMode())

	// After 3 seconds, command timestamp exceeds threshold, switch to command mode
	time.Sleep(2 * time.Second)
	assert.True(t, svc.InCommandMode())

	// After 4 seconds, command timestamp still exceeds threshold, remain in command mode
	time.Sleep(time.Second)
	assert.True(t, svc.InCommandMode())

	encPayload, _ := json.Marshal(map[string]interface{}{"action": "fromTest", "value": 23})
	token := mqttClient.Publish(cfg.MQTT.ReadCommandTopic, 1, false, encPayload)
	token.Wait()
	testLogger.Info("Test published new cloud cmd to", "topic", cfg.MQTT.ReadCommandTopic)

	// After 5 seconds, new command received, resume standby mode
	time.Sleep(time.Second)
	assert.True(t, svc.InStandbyMode())

	svc.Stop()
}
