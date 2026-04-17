package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"confirmate.io/collectors/cloud"
)

func main() {
	var err error

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := cloud.LoadConfigFromEnv()
	if err != nil {
		logger.Error("invalid configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	runCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err = cloud.Start(runCtx, cfg, logger)
	if err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("collector stopped with error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
