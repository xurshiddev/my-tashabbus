package handlers

import (
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/my-tashabbus/bot/internal/telegram/keyboards"
)

const WelcomeMessage = "Assalomu alaykum! My Tashabbus tizimiga xush kelibsiz. Mini App orqali tizimga kirish va o'zingizga biriktirilgan ko'chalarni ko'rish uchun avval ro'yxatdan o'tgan bo'lishingiz kerak."

type MessageSender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

type StartHandler struct {
	sender     MessageSender
	miniAppURL string
	log        *slog.Logger
}

func NewStartHandler(sender MessageSender, miniAppURL string, log *slog.Logger) StartHandler {
	return StartHandler{sender: sender, miniAppURL: miniAppURL, log: log}
}

func (h StartHandler) Handle(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}
	message := tgbotapi.NewMessage(update.Message.Chat.ID, WelcomeMessage)
	message.ReplyMarkup = keyboards.MiniAppKeyboard(h.miniAppURL)

	if _, err := h.sender.Send(message); err != nil {
		h.log.Error("send start message", "error", err)
	}
}
