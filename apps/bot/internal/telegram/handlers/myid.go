package handlers

import (
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const MyIDMessageTemplate = "Sizning Telegram ID: %d\nUsername: %s"

type MyIDHandler struct {
	sender MessageSender
	log    *slog.Logger
}

func NewMyIDHandler(sender MessageSender, log *slog.Logger) MyIDHandler {
	return MyIDHandler{sender: sender, log: log}
}

func FormatMyIDMessage(user *tgbotapi.User) string {
	if user == nil {
		return "Sizning Telegram ID: \nUsername: "
	}
	return fmt.Sprintf(MyIDMessageTemplate, user.ID, user.UserName)
}

func (h MyIDHandler) Handle(update tgbotapi.Update) {
	if update.Message == nil || update.Message.From == nil {
		return
	}

	message := tgbotapi.NewMessage(update.Message.Chat.ID, FormatMyIDMessage(update.Message.From))
	if _, err := h.sender.Send(message); err != nil {
		h.log.Error("send myid message", "error", err)
	}
}
