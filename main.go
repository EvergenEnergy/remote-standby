package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/EvergenEnergy/remote-standby/config"
	"github.com/EvergenEnergy/remote-standby/standby"
	"github.com/EvergenEnergy/remote-standby/worker"
)

var logLevels = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

func main() {
	cfg, err := config.FromEnv()
	if err != nil {
		log.Fatalf("reading config: %s", err)
	}

	cfgLevel, exists := logLevels[strings.ToLower(cfg.Logging.Level)]
	if !exists {
		cfgLevel = slog.LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: cfgLevel}))

	standbyService := standby.NewService(logger, cfg)
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
