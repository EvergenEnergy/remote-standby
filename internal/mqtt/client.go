package mqtt

import (
	"fmt"

	"github.com/EvergenEnergy/remote-standby/internal/config"
	pahoMQTT "github.com/eclipse/paho.mqtt.golang"
)

type ClientWrapper struct {
	client pahoMQTT.Client
}

func (c *ClientWrapper) Subscribe(topic string, qos byte, handler pahoMQTT.MessageHandler) MqttToken {
	return c.client.Subscribe(topic, qos, handler)
}

func (c *ClientWrapper) Connect() MqttToken {
	return c.client.Connect()
}

func (c *ClientWrapper) Disconnect(delay uint) {
	c.client.Disconnect(delay)
}

func (c *ClientWrapper) Publish(topic string, qos byte, retained bool, payload interface{}) MqttToken {
	return c.client.Publish(topic, qos, retained, payload)
}

var onConnect pahoMQTT.OnConnectHandler = func(client pahoMQTT.Client) {
	fmt.Println("Connected")
}

var onConnectionLost pahoMQTT.ConnectionLostHandler = func(client pahoMQTT.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func NewClient(cfg config.Config) *ClientWrapper {
	brokerURL := cfg.MQTT.BrokerURL

	mqttOpts := pahoMQTT.NewClientOptions()
	mqttOpts.AddBroker(brokerURL)
	// Note we don't set a ClientID here, the AWS bridge handles that
	mqttOpts.SetCleanSession(true)
	mqttOpts.SetAutoReconnect(true)
	mqttOpts.SetOrderMatters(true)
	mqttOpts.SetOnConnectHandler(onConnect)
	mqttOpts.SetConnectionLostHandler(onConnectionLost)

	return &ClientWrapper{client: pahoMQTT.NewClient(mqttOpts)}
}
