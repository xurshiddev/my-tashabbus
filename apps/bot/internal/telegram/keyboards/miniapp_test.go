package keyboards

import (
	"encoding/json"
	"testing"
)

func TestMiniAppKeyboardWithURL(t *testing.T) {
	keyboard := MiniAppKeyboard("https://example.com/miniapp")
	if len(keyboard.InlineKeyboard) != 1 {
		t.Fatalf("expected one keyboard row, got %d", len(keyboard.InlineKeyboard))
	}
	button := keyboard.InlineKeyboard[0][0]
	if button.Text != MiniAppButtonText {
		t.Fatalf("expected button text %q, got %q", MiniAppButtonText, button.Text)
	}
	if button.WebApp == nil || button.WebApp.URL != "https://example.com/miniapp" {
		t.Fatalf("expected web app URL to be set")
	}
	payload, err := json.Marshal(keyboard)
	if err != nil {
		t.Fatalf("marshal keyboard: %v", err)
	}

	var decoded struct {
		InlineKeyboard [][]map[string]any `json:"inline_keyboard"`
	}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal keyboard: %v", err)
	}
	buttonPayload := decoded.InlineKeyboard[0][0]
	if _, ok := buttonPayload["url"]; ok {
		t.Fatalf("expected no root URL button field, got %s", string(payload))
	}
	webApp, ok := buttonPayload["web_app"].(map[string]any)
	if !ok {
		t.Fatalf("expected web_app button payload, got %s", string(payload))
	}
	if webApp["url"] != "https://example.com/miniapp" {
		t.Fatalf("expected web_app URL, got %v", webApp["url"])
	}
}

func TestMiniAppKeyboardWithoutURL(t *testing.T) {
	keyboard := MiniAppKeyboard("")
	if len(keyboard.InlineKeyboard) != 0 {
		t.Fatalf("expected no keyboard rows, got %d", len(keyboard.InlineKeyboard))
	}
}
