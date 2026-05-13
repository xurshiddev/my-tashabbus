package app

import (
	"context"
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/my-tashabbus/bot/internal/config"
	"github.com/my-tashabbus/bot/internal/telegram/handlers"
)

type App struct {
	cfg config.Config
	log *slog.Logger
	bot *tgbotapi.BotAPI
}

func New(cfg config.Config, log *slog.Logger) (*App, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, fmt.Errorf("initialize telegram api: %w", err)
	}
	return &App{cfg: cfg, log: log, bot: bot}, nil
}

func (a *App) Run(ctx context.Context) error {
	a.log.Info("telegram bot starting", "username", a.bot.Self.UserName)

	updates := a.bot.GetUpdatesChan(tgbotapi.NewUpdate(0))
	defer a.bot.StopReceivingUpdates()

	startHandler := handlers.NewStartHandler(a.bot, a.cfg.MiniAppURL, a.log)

	for {
		select {
		case <-ctx.Done():
			a.log.Info("telegram bot shutdown signal received")
			return ctx.Err()
		case update, ok := <-updates:
			if !ok {
				return nil
			}
			if update.Message == nil || !update.Message.IsCommand() {
				continue
			}
			if update.Message.Command() == "start" {
				startHandler.Handle(update)
			}
		}
	}
}
