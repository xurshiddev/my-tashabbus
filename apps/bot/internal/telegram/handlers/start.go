package handlers

import (
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/my-tashabbus/bot/internal/telegram/keyboards"
)

const WelcomeMessage = "Assalomu alaykum! My Tashabbus tizimiga xush kelibsiz. Mini App orqali o'zingizga biriktirilgan ko'chalar va xonadonlarni ko'rishingiz mumkin."
const ChairmanWelcomeMessage = "Assalomu alaykum, MFY raisi! Mini App orqali MFY ko'chalari, xonadonlari va mas'ullarini boshqarishingiz mumkin."
const UnassignedWelcomeMessage = "Assalomu alaykum! Agar tizimga biriktirilmagan bo'lsangiz, /myid orqali Telegram ID ni olib MFY raisiga yuboring."

type MessageSender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

type StartHandler struct {
	sender     MessageSender
	miniAppURL string
	ownerID    int64
	log        *slog.Logger
}

func NewStartHandler(sender MessageSender, miniAppURL string, ownerID int64, log *slog.Logger) StartHandler {
	return StartHandler{sender: sender, miniAppURL: miniAppURL, ownerID: ownerID, log: log}
}

func (h StartHandler) Handle(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}
	message := tgbotapi.NewMessage(update.Message.Chat.ID, h.welcomeText(update.Message.From))
	message.ReplyMarkup = keyboards.MiniAppKeyboard(h.miniAppURL)

	if _, err := h.sender.Send(message); err != nil {
		h.log.Error("send start message", "error", err)
	}
}

func (h StartHandler) welcomeText(user *tgbotapi.User) string {
	if user != nil && h.ownerID != 0 && user.ID == h.ownerID {
		return ChairmanWelcomeMessage
	}
	return WelcomeMessage + "\n\n" + UnassignedWelcomeMessage
}
