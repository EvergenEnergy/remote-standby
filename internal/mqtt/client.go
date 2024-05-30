package client

import (
	"fmt"

	"github.com/EvergenEnergy/remote-standby/internal/config"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var onConnect mqtt.OnConnectHandler = func(_ mqtt.Client) {
	fmt.Println("Connected")
}

var onConnectionLost mqtt.ConnectionLostHandler = func(_ mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func NewClient(cfg config.Config) mqtt.Client {
	brokerURL := cfg.MQTT.BrokerURL

	mqttOpts := mqtt.NewClientOptions()
	mqttOpts.AddBroker(brokerURL)
	// Note we don't set a ClientID here, the AWS bridge handles that
	mqttOpts.SetCleanSession(true)
	mqttOpts.SetAutoReconnect(true)
	mqttOpts.SetOrderMatters(true)
	mqttOpts.SetOnConnectHandler(onConnect)
	mqttOpts.SetConnectionLostHandler(onConnectionLost)

	return mqtt.NewClient(mqttOpts)
}
