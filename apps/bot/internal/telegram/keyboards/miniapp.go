package keyboards

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

const MiniAppButtonText = "Mini Appni ochish"

func MiniAppKeyboard(miniAppURL string) tgbotapi.InlineKeyboardMarkup {
	if miniAppURL == "" {
		return tgbotapi.NewInlineKeyboardMarkup()
	}

	button := tgbotapi.NewInlineKeyboardButtonURL(MiniAppButtonText, miniAppURL)
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(button))
}
