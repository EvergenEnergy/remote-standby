package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/EvergenEnergy/remote-standby/internal/config"
	internalMQTT "github.com/EvergenEnergy/remote-standby/internal/mqtt"
	"github.com/EvergenEnergy/remote-standby/internal/outagelog"
	"github.com/EvergenEnergy/remote-standby/internal/publisher"
	"github.com/EvergenEnergy/remote-standby/internal/standby"
	"github.com/EvergenEnergy/remote-standby/internal/storage"
	"github.com/EvergenEnergy/remote-standby/internal/worker"
)

var logLevels = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

func main() {
	cfg, err := config.FromFile()
	if err != nil {
		log.Fatalf("reading config: %s", err)
	}

	cfgLevel, exists := logLevels[strings.ToLower(cfg.Logging.Level)]
	if !exists {
		cfgLevel = slog.LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: cfgLevel}))

	logHandle, err := outagelog.Open(cfg.Standby.OutageLogFile)
	if err != nil {
		logger.Error("Could not open logfile", "path", cfg.Standby.OutageLogFile)
	}

	logHandler := outagelog.NewHandler(logHandle, logger)
	defer logHandler.Close()

	mqttClient := internalMQTT.NewClient(cfg)
	storageService := storage.NewService(logger)
	publisher := publisher.NewService(logger, cfg, mqttClient)
	standbyService := standby.NewService(logger, cfg, storageService, publisher, logHandler, mqttClient)
	standbyWorker := worker.NewWorker(logger, cfg, standbyService)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Interrupt)
	defer cancel()

	go func() {
		err := standbyWorker.Start(ctx)
		if err != nil {
			panic(err)
		}
	}()

	<-ctx.Done()

	_ = standbyWorker.Stop()
}
