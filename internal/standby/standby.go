package standby

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/EvergenEnergy/remote-standby/internal/config"
	mqtt "github.com/EvergenEnergy/remote-standby/internal/mqtt"
	"github.com/EvergenEnergy/remote-standby/internal/plan"
	"github.com/EvergenEnergy/remote-standby/internal/storage"
	pahoMQTT "github.com/eclipse/paho.mqtt.golang"
)

type ServiceMode string

const (
	StandbyMode ServiceMode = "standby"
	CommandMode ServiceMode = "command"
)

type Service struct {
	logger     *slog.Logger
	cfg        config.Config
	client     mqtt.MqttClient
	storageSvc *storage.Service
	mutex      *sync.Mutex
	mode       ServiceMode
}

func NewService(logger *slog.Logger, cfg config.Config, storage *storage.Service, client mqtt.MqttClient) *Service {
	return &Service{
		logger:     logger,
		cfg:        cfg,
		client:     client,
		storageSvc: storage,
		mutex:      new(sync.Mutex),
		mode:       StandbyMode,
	}
}

func (s *Service) subscribeToTopic(topic string, handler pahoMQTT.MessageHandler) {
	token := s.client.Subscribe(topic, 1, handler)
	token.Wait()
	s.logger.Debug("Subscribed to topic " + topic)
}

func (s *Service) handleCommandMessage(client pahoMQTT.Client, msg pahoMQTT.Message) {
	s.logger.Debug(fmt.Sprintf("Received message: %s from topic: %s", msg.Payload(), msg.Topic()))
	s.storageSvc.SetCommandTimestamp(time.Now())
}

func (s *Service) handlePlanMessage(client pahoMQTT.Client, msg pahoMQTT.Message) {
	s.logger.Debug(fmt.Sprintf("Received plan in message: %s from topic: %s", msg.Payload(), msg.Topic()))
	optPlan := plan.OptimisationPlan{}

	err := json.Unmarshal(msg.Payload(), &optPlan)
	if err == nil && optPlan.IsEmpty(s.logger) {
		err = fmt.Errorf("optimisation plan is empty")
	}
	if err != nil {
		s.publishError("reading optimisation plan", err)
	}

	handler := plan.NewHandler(s.cfg.Standby.BackupFile)
	err = handler.WritePlan(optPlan)
	if err != nil {
		s.publishError("writing optimisation plan", err)
	}
}

func (s *Service) runMQTT() error {
	if token := s.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	s.subscribeToTopic(s.cfg.MQTT.StandbyTopic, s.handlePlanMessage)
	s.subscribeToTopic(s.cfg.MQTT.ReadCommandTopic, s.handleCommandMessage)

	return nil
}

func (s *Service) stopMQTT() {
	s.client.Disconnect(uint(1000))
}

func (s *Service) runDetector(ctx context.Context) {
	checkInterval := s.cfg.Standby.CheckInterval

	ticker := time.NewTicker(checkInterval)
	for {
		select {
		case <-ticker.C:
			s.CheckForOutage(time.Now())
		case <-ctx.Done():
			ticker.Stop()

			s.logger.Info("Shutting down detector")

			return
		}
	}
}

func (s *Service) CheckForOutage(currentTime time.Time) {
	outageThreshold := s.cfg.Standby.OutageThreshold
	timeSinceLastCmd := currentTime.Sub(s.storageSvc.GetCommandTimestamp())
	s.logger.Debug("checking", "time since last command", timeSinceLastCmd)
	if timeSinceLastCmd > outageThreshold {
		if s.InStandbyMode() {
			s.logger.Info("Outage detected", "config threshold", outageThreshold, "time since last command", timeSinceLastCmd)
			s.setMode(CommandMode)
		} else {
			s.logger.Debug("Ongoing outage", "config threshold", outageThreshold, "time since last command", timeSinceLastCmd)
		}
		// TODO: Trigger a process to identify closest optimisation command from plan and send it
	} else {
		if s.InCommandMode() {
			s.logger.Info("Commands resumed after outage", "time since last command", timeSinceLastCmd)
			s.setMode(StandbyMode)
		}
	}
}

func (s *Service) Start(ctx context.Context) error {
	if err := s.runMQTT(); err != nil {
		return fmt.Errorf("starting MQTT client: %w", err)
	}
	go s.runDetector(ctx)
	return nil
}

func (s *Service) Stop() {
	s.stopMQTT()
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

	s.client.Publish(errTopic, 1, false, encPayload)
}

func (s *Service) setMode(newMode ServiceMode) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.mode = newMode
}

func (s *Service) getMode() ServiceMode {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.mode
}

func (s *Service) InStandbyMode() bool {
	return s.getMode() == StandbyMode
}

func (s *Service) InCommandMode() bool {
	return s.getMode() == CommandMode
}
