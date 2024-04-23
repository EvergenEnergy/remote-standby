package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/EvergenEnergy/remote-standby/config"
	"github.com/EvergenEnergy/remote-standby/standby"
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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	standbyWorker := standby.Init(logger, cfg)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		standbyWorker.Start(ctx)
	}()

	wg.Wait()
	_ = standbyWorker.Stop()
}
