package standby

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/EvergenEnergy/remote-standby/internal/config"
	"github.com/EvergenEnergy/remote-standby/internal/plan"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type ServiceMode string

var (
	StandbyMode = ServiceMode("standby")
	CommandMode = ServiceMode("command")
)

type Service struct {
	logger                *slog.Logger
	cfg                   config.Config
	MQTTClient            mqtt.Client
	LatestCommandReceived time.Time
	Mode                  ServiceMode
	mutex                 sync.Mutex
}

func NewService(logger *slog.Logger, cfg config.Config) *Service {
	return &Service{
		logger:     logger,
		cfg:        cfg,
		MQTTClient: initMQTTClient(cfg),
		Mode:       StandbyMode,
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
	mqttOpts.SetCleanSession(true)
	mqttOpts.SetAutoReconnect(true)
	mqttOpts.SetOrderMatters(true)
	mqttOpts.SetOnConnectHandler(onConnect)
	mqttOpts.SetConnectionLostHandler(onConnectionLost)

	return mqtt.NewClient(mqttOpts)
}

func (s *Service) subscribeToTopic(topic string, handler mqtt.MessageHandler) {
	token := s.MQTTClient.Subscribe(topic, 1, handler)
	token.Wait()
	s.logger.Debug("Subscribed to topic " + topic)
}

func (s *Service) RunMQTT(ctx context.Context) error {
	if token := s.MQTTClient.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	s.subscribeToTopic(s.cfg.MQTT.StandbyTopic,
		func(client mqtt.Client, msg mqtt.Message) {
			s.logger.Debug(fmt.Sprintf("Received message: %s from topic: %s", msg.Payload(), msg.Topic()))
			optPlan := plan.OptimisationPlan{}

			err := json.Unmarshal(msg.Payload(), &optPlan)
			if err != nil {
				s.publishError("reading optimisation plan", err)
			}
			handler := plan.NewHandler(s.cfg.Standby.BackupFile)
			err = handler.WritePlan(optPlan)
			if err != nil {
				s.publishError("writing optimisation plan", err)
			}
		})

	s.subscribeToTopic(s.cfg.MQTT.ReadCommandTopic,
		func(client mqtt.Client, msg mqtt.Message) {
			s.logger.Debug(fmt.Sprintf("Received message: %s from topic: %s", msg.Payload(), msg.Topic()))
			s.setCommandTimestamp()
		})

	return nil
}

func (s *Service) StopMQTT() {
	s.MQTTClient.Disconnect(uint(1000))
}

func (s *Service) RunDetector(ctx context.Context) {
	checkInterval := time.Duration(s.cfg.Standby.CheckInterval) * time.Second
	outageDuration := time.Duration(s.cfg.Standby.OutageInterval) * time.Second

	// initialise this so we don't immediately go into failure mode on startup
	s.setCommandTimestamp()

	ticker := time.NewTicker(checkInterval)
	for {
		select {
		case <-ticker.C:
			s.CheckForOutage(outageDuration)
		case <-ctx.Done():
			ticker.Stop()

			s.logger.Info("Shutting down detector")

			return
		}
	}
}

func (s *Service) CheckForOutage(outageDuration time.Duration) {
	timeSinceLastCmd := time.Since(s.getCommandTimestamp())
	s.logger.Debug("checking", "time since last command", timeSinceLastCmd)
	if timeSinceLastCmd > outageDuration {
		if s.Mode == StandbyMode {
			s.logger.Info("Outage detected", "config duration", outageDuration, "time since last command", timeSinceLastCmd)
			s.Mode = CommandMode
		} else {
			s.logger.Debug("Ongoing outage", "config duration", outageDuration, "time since last command", timeSinceLastCmd)
		}
		// TODO: Trigger a process to identify closest optimisation command from plan and send it
	} else {
		if s.Mode == CommandMode {
			s.logger.Info("Commands resumed after outage", "time since last command", timeSinceLastCmd)
			s.Mode = StandbyMode
		}
	}
}

func (s *Service) Start(ctx context.Context) error {
	if err := s.RunMQTT(ctx); err != nil {
		return fmt.Errorf("starting MQTT client: %w", err)
	}
	go s.RunDetector(ctx)
	return nil
}

func (s *Service) Stop() {
	s.StopMQTT()
}

type ErrorPayload struct {
	Category  string    `json:"category"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

func (s *Service) publishError(message string, receivedError error) {
	s.logger.Error(message, "error", receivedError)

	payload := ErrorPayload{
		Category:  "Standby",
		Message:   fmt.Sprintf("Error %s: %s", message, receivedError),
		Timestamp: time.Now(),
	}
	encPayload, err := json.Marshal(payload)
	if err != nil {
		s.logger.Error("marshalling error payload", "error", err)
	}

	errTopic := fmt.Sprintf("%s/%s", s.cfg.MQTT.ErrorTopic, payload.Category)

	s.MQTTClient.Publish(errTopic, 1, false, encPayload)
}

func (s *Service) setCommandTimestamp() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.LatestCommandReceived = time.Now()
}

func (s *Service) getCommandTimestamp() time.Time {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.LatestCommandReceived
}
