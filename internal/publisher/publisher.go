package publisher

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/EvergenEnergy/remote-standby/internal/config"
	"github.com/EvergenEnergy/remote-standby/internal/plan"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Service struct {
	logger     *slog.Logger
	cfg        config.Config
	mqttClient mqtt.Client
}

func NewService(logger *slog.Logger, cfg config.Config, mqttClient mqtt.Client) *Service {
	return &Service{
		logger:     logger,
		cfg:        cfg,
		mqttClient: mqttClient,
	}
}

type CommandPayload struct {
	Action string  `json:"action"`
	Value  float64 `json:"value"`
}

type ErrorPayload struct {
	Category  string    `json:"category"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

const errorCategory = "Standby"

func (s *Service) PublishError(message string, receivedError error) {
	s.logger.Error(message, "error", receivedError)

	payload := ErrorPayload{
		Category:  errorCategory,
		Message:   fmt.Sprintf("Error %s: %s", message, receivedError),
		Timestamp: time.Now(),
	}
	encPayload, err := json.Marshal(payload)
	if err != nil {
		s.logger.Error("marshalling error payload", "error", err)
		return
	}

	if s.cfg.MQTT.ErrorTopic == "" {
		s.logger.Error("no error topic configured")
		return
	}
	errTopic := fmt.Sprintf("%s/%s", s.cfg.MQTT.ErrorTopic, payload.Category)

	s.mqttClient.Publish(errTopic, 1, false, encPayload)
}

func (s *Service) PublishCommand(optInterval plan.OptimisationInterval) error {
	if s.cfg.MQTT.WriteCommandTopic == "" || s.cfg.MQTT.CommandAction == "" {
		return fmt.Errorf("no command topic (%s) or action (%s) configured", s.cfg.MQTT.WriteCommandTopic, s.cfg.MQTT.CommandAction)
	}

	payload := BuildCommandPayload(s.cfg.MQTT.CommandAction, optInterval)
	encPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshalling command payload: %w", err)
	}

	s.mqttClient.Publish(s.cfg.MQTT.WriteCommandTopic, 1, false, encPayload)
	return nil
}

func BuildCommandPayload(action string, optInterval plan.OptimisationInterval) CommandPayload {
	meterValue := float64(optInterval.MeterPower.Value)
	meterUnit := optInterval.MeterPower.Unit

	var publishMeterValue float64
	switch {
	case meterUnit == 1:
		publishMeterValue = meterValue / 1000
	case meterUnit == 2:
		publishMeterValue = meterValue
	case meterUnit == 3:
		publishMeterValue = meterValue * 1000
	}
	return CommandPayload{
		Action: action,
		Value:  publishMeterValue,
	}
}
