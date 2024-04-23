package standby

import (
	"context"
	"log/slog"

	"github.com/EvergenEnergy/remote-standby/config"
)

type Worker struct {
	logger *slog.Logger
	cfg    config.Config
}

func Init(logger *slog.Logger, cfg config.Config) *Worker {
	logger.Info("initing worker")
	return &Worker{
		logger: logger,
		cfg:    cfg,
	}
}

func (w *Worker) Start(_ context.Context) {
	w.logger.Info("started worker")
}

func (w *Worker) Stop() error {
	w.logger.Info("stopping worker")
	return nil
}
