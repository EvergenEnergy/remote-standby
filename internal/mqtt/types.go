package mqtt

import (
	"time"

	pahoMQTT "github.com/eclipse/paho.mqtt.golang"
)

type MqttToken interface {
	Wait() bool
	Error() error
	Done() <-chan struct{}
	WaitTimeout(time.Duration) bool
}

type MqttClient interface {
	Subscribe(topic string, qos byte, handler pahoMQTT.MessageHandler) MqttToken
	Connect() MqttToken
	Disconnect(delay uint)
	Publish(topic string, qos byte, retained bool, payload interface{}) MqttToken
}
