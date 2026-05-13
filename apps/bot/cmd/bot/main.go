package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/my-tashabbus/bot/internal/app"
	"github.com/my-tashabbus/bot/internal/config"
	"github.com/my-tashabbus/bot/internal/logger"
)

func main() {
	if err := run(); err != nil {
		slog.Error("bot stopped with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()
	log := logger.New(cfg.BotEnv)

	if cfg.BotToken == "" {
		if cfg.BotEnv == "production" {
			return config.ErrMissingBotToken
		}
		log.Warn("BOT_TOKEN is empty; bot service will idle until stopped")
		waitForShutdown(context.Background(), log)
		return nil
	}

	botApp, err := app.New(cfg, log)
	if err != nil {
		return fmt.Errorf("create bot app: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := botApp.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("run bot app: %w", err)
	}
	return nil
}

func waitForShutdown(parent context.Context, log *slog.Logger) {
	ctx, stop := signal.NotifyContext(parent, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	log.Info("bot shutdown signal received")
}
