package keyboards

const MiniAppButtonText = "Mini Appni ochish"

type WebAppInfo struct {
	URL string `json:"url"`
}

type WebAppInlineKeyboardButton struct {
	Text   string      `json:"text"`
	WebApp *WebAppInfo `json:"web_app,omitempty"`
}

type WebAppInlineKeyboardMarkup struct {
	InlineKeyboard [][]WebAppInlineKeyboardButton `json:"inline_keyboard"`
}

func MiniAppKeyboard(miniAppURL string) WebAppInlineKeyboardMarkup {
	if miniAppURL == "" {
		return WebAppInlineKeyboardMarkup{}
	}

	button := WebAppInlineKeyboardButton{
		Text:   MiniAppButtonText,
		WebApp: &WebAppInfo{URL: miniAppURL},
	}
	return WebAppInlineKeyboardMarkup{InlineKeyboard: [][]WebAppInlineKeyboardButton{{button}}}
}
