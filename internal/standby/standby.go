package standby

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/EvergenEnergy/remote-standby/internal/config"
	"github.com/EvergenEnergy/remote-standby/internal/outagelog"
	"github.com/EvergenEnergy/remote-standby/internal/plan"
	"github.com/EvergenEnergy/remote-standby/internal/publisher"
	"github.com/EvergenEnergy/remote-standby/internal/storage"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type ServiceMode string

const (
	StandbyMode ServiceMode = "standby"
	CommandMode ServiceMode = "command"
)

type Service struct {
	logger      *slog.Logger
	cfg         config.Config
	mqttClient  mqtt.Client
	storageSvc  *storage.Service
	publisher   *publisher.Service
	planHandler plan.Handler
	mutex       *sync.Mutex
	mode        ServiceMode
	logHandler  *outagelog.Handler
}

func NewService(
	logger *slog.Logger, cfg config.Config, storage *storage.Service,
	publisher *publisher.Service, logHandler *outagelog.Handler, mqttClient mqtt.Client,
) *Service {
	return &Service{
		logger:      logger,
		cfg:         cfg,
		mqttClient:  mqttClient,
		storageSvc:  storage,
		publisher:   publisher,
		mutex:       new(sync.Mutex),
		mode:        StandbyMode,
		planHandler: plan.NewHandler(logger, cfg.Standby.BackupFile),
		logHandler:  logHandler,
	}
}

func (s *Service) subscribeToTopic(topic string, handler mqtt.MessageHandler) {
	token := s.mqttClient.Subscribe(topic, 1, handler)
	token.Wait()
	s.logger.Debug("Subscribed to topic " + topic)
}

func (s *Service) handleCommandMessage(_ mqtt.Client, msg mqtt.Message) {
	s.logger.Debug(fmt.Sprintf("Received command: %s from topic: %s", msg.Payload(), msg.Topic()))
	s.storageSvc.SetCommandTimestamp(time.Now())
}

func (s *Service) handlePlanMessage(_ mqtt.Client, msg mqtt.Message) {
	s.logger.Debug(fmt.Sprintf("Received plan in message: %s from topic: %s", msg.Payload(), msg.Topic()))

	optPlan := plan.OptimisationPlan{}

	err := json.Unmarshal(msg.Payload(), &optPlan)
	if err == nil && optPlan.IsEmpty() {
		err = fmt.Errorf("optimisation plan is empty")
	}
	if err != nil {
		s.publisher.PublishError("reading optimisation plan", err)
	}

	err = s.planHandler.WritePlan(optPlan)
	if err != nil {
		s.publisher.PublishError("writing optimisation plan", err)
	}
}

func (s *Service) runMQTT() error {
	if token := s.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("mqtt token: %w", token.Error())
	}

	s.subscribeToTopic(s.cfg.MQTT.StandbyTopic, s.handlePlanMessage)
	s.subscribeToTopic(s.cfg.MQTT.ReadCommandTopic, s.handleCommandMessage)

	return nil
}

func (s *Service) stopMQTT() {
	s.mqttClient.Disconnect(uint(1000))
}

func (s *Service) runDetector(ctx context.Context) {
	checkInterval := s.cfg.Standby.CheckInterval

	ticker := time.NewTicker(checkInterval)

	for {
		select {
		case <-ticker.C:
			s.checkForOutage(time.Now())
		case <-ctx.Done():
			ticker.Stop()

			s.logger.Info("Shutting down detector")

			return
		}
	}
}

func (s *Service) checkForOutage(currentTime time.Time) {
	outageThreshold := s.cfg.Standby.OutageThreshold
	timeSinceLastCmd := currentTime.Sub(s.storageSvc.GetCommandTimestamp())
	s.logger.Debug("checking", "time since last command", timeSinceLastCmd, "current mode", s.getMode(), "currentTime", currentTime)

	if timeSinceLastCmd < outageThreshold {
		if s.InCommandMode() {
			s.logger.Info("Commands resumed after outage", "time since last command", timeSinceLastCmd)
			s.setMode(StandbyMode)
			s.logHandler.Append("Resumed standby mode", map[string]string{"timeSinceLastCmd": timeSinceLastCmd.String()})
		}
		return
	}

	if s.InStandbyMode() {
		s.logger.Info("Outage detected", "config threshold", outageThreshold, "time since last command", timeSinceLastCmd)
		s.setMode(CommandMode)
		s.logHandler.Append("Entered command mode", map[string]string{"timeSinceLastCmd": timeSinceLastCmd.String()})
	}

	currentInterval, err := s.planHandler.GetCurrentInterval(currentTime)
	if err != nil {
		s.logHandler.Append("No command available", nil)

		s.publisher.PublishError("getting current command", err)
		return
	}

	err = s.publisher.PublishCommand(currentInterval)
	if err != nil {
		s.publisher.PublishError("publishing current command", err)
	}

	s.logHandler.Append("Published command", currentInterval.LogFormat())
}

func (s *Service) Start(ctx context.Context) error {
	if err := s.runMQTT(); err != nil {
		return fmt.Errorf("starting MQTT client: %w", err)
	}

	go s.runDetector(ctx)
	s.logHandler.Append("Service started", nil)
	return nil
}

func (s *Service) Stop() {
	s.stopMQTT()
	s.logHandler.Append("Service stopped", nil)
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
