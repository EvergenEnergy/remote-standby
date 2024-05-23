package mqtt

import (
	"time"

	pahoMQTT "github.com/eclipse/paho.mqtt.golang"
)

type TokenWrapper struct {
	token pahoMQTT.Token
}

func (w *TokenWrapper) Wait() bool {
	return w.token.Wait()
}

func (w *TokenWrapper) WaitTimeout(duration time.Duration) bool {
	return w.token.WaitTimeout(duration)
}

func (w *TokenWrapper) Error() error {
	return w.token.Error()
}

func (w *TokenWrapper) Done() <-chan struct{} {
	return w.token.Done()
}
