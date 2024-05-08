package standby

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/EvergenEnergy/remote-standby/config"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Worker struct {
	logger     *slog.Logger
	cfg        config.Config
	mqttClient mqtt.Client
}

func Init(logger *slog.Logger, cfg config.Config) *Worker {
	return &Worker{
		logger:     logger,
		cfg:        cfg,
		mqttClient: initMQTTClient(cfg, logger),
	}
}

func (w *Worker) Start(ctx context.Context) error {
	err := w.RunMQTT(ctx)
	if err != nil {
		return fmt.Errorf("running mqtt: %w", err)
	}
	return nil
}

func (w *Worker) Stop() error {
	return nil
}

var onConnect mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var onConnectionLost mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func initMQTTClient(cfg config.Config, logger *slog.Logger) mqtt.Client {
	brokerURL := cfg.MQTT.BrokerURL

	mqttOpts := mqtt.NewClientOptions()
	mqttOpts.AddBroker(brokerURL)
	mqttOpts.SetClientID("remote-standby-client")
	mqttOpts.SetCleanSession(true)
	mqttOpts.SetAutoReconnect(true)
	mqttOpts.SetOrderMatters(true)
	mqttOpts.SetOnConnectHandler(onConnect)
	mqttOpts.SetConnectionLostHandler(onConnectionLost)

	return mqtt.NewClient(mqttOpts)
}

func (w *Worker) subscribeToTopic(topic string) {
	token := w.mqttClient.Subscribe(topic, 1,
		func(client mqtt.Client, msg mqtt.Message) {
			w.logger.Debug(fmt.Sprintf("Received message: %s from topic: %s", msg.Payload(), msg.Topic()))
		})
	token.Wait()
	w.logger.Debug("Subscribed to topic " + topic)
}

func (w *Worker) RunMQTT(ctx context.Context) error {
	if token := w.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	w.subscribeToTopic(w.cfg.MQTT.StandbyTopic)

	// wait until the context is cancelled
	<-ctx.Done()

	w.mqttClient.Disconnect(uint(1000))

	return nil
}
