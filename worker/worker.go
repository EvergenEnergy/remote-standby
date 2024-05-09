package worker

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/EvergenEnergy/remote-standby/config"
	"github.com/EvergenEnergy/remote-standby/standby"
)

type Worker struct {
	logger     *slog.Logger
	cfg        config.Config
	standbySvc *standby.Service
}

func NewWorker(logger *slog.Logger, cfg config.Config, standby *standby.Service) *Worker {
	return &Worker{
		logger:     logger,
		cfg:        cfg,
		standbySvc: standby,
	}
}

func (w *Worker) Start(ctx context.Context) error {
	err := w.standbySvc.RunMQTT(ctx)
	if err != nil {
		return fmt.Errorf("running standby service: %w", err)
	}
	return nil
}

func (w *Worker) Stop() error {
	w.standbySvc.StopMQTT()
	return nil
}
