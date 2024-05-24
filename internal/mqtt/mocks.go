package mqtt

import (
	"time"

	pahoMQTT "github.com/eclipse/paho.mqtt.golang"
)

type MockToken struct{}

func (w *MockToken) Wait() bool {
	return true
}

func (w *MockToken) WaitTimeout(delay time.Duration) bool {
	return true
}

func (w *MockToken) Error() error {
	return nil
}

func (w *MockToken) Done() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

type MockClient struct {
	SubscribedTopics []string
}

func (m *MockClient) Subscribe(topic string, qos byte, handler pahoMQTT.MessageHandler) MqttToken {
	m.SubscribedTopics = append(m.SubscribedTopics, topic)
	token := &MockToken{}
	return &TokenWrapper{token: token}
}

func (m *MockClient) Connect() MqttToken {
	token := &MockToken{}
	return &TokenWrapper{token: token}
}

func (m *MockClient) Disconnect(delay uint) {}

func (m *MockClient) Publish(topic string, qos byte, retained bool, payload interface{}) MqttToken {
	token := &MockToken{}
	return &TokenWrapper{token: token}
}
