package standby

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/EvergenEnergy/remote-standby/config"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Service struct {
	logger     *slog.Logger
	cfg        config.Config
	MQTTClient mqtt.Client
}

func NewService(logger *slog.Logger, cfg config.Config) *Service {
	return &Service{
		logger:     logger,
		cfg:        cfg,
		MQTTClient: initMQTTClient(cfg),
	}
}

var onConnect mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var onConnectionLost mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func initMQTTClient(cfg config.Config) mqtt.Client {
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

func (s *Service) subscribeToTopic(topic string) {
	token := s.MQTTClient.Subscribe(topic, 1,
		func(client mqtt.Client, msg mqtt.Message) {
			s.logger.Debug(fmt.Sprintf("Received message: %s from topic: %s", msg.Payload(), msg.Topic()))
		})
	token.Wait()
	s.logger.Debug("Subscribed to topic " + topic)
}

func (s *Service) RunMQTT(ctx context.Context) error {
	if token := s.MQTTClient.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	s.subscribeToTopic(s.cfg.MQTT.StandbyTopic)

	return nil
}

func (s *Service) StopMQTT() {
	s.MQTTClient.Disconnect(uint(1000))
}
