package keyboards

import "testing"

func TestMiniAppKeyboardWithURL(t *testing.T) {
	keyboard := MiniAppKeyboard("https://example.com/miniapp")
	if len(keyboard.InlineKeyboard) != 1 {
		t.Fatalf("expected one keyboard row, got %d", len(keyboard.InlineKeyboard))
	}
	button := keyboard.InlineKeyboard[0][0]
	if button.Text != MiniAppButtonText {
		t.Fatalf("expected button text %q, got %q", MiniAppButtonText, button.Text)
	}
	if button.URL == nil || *button.URL != "https://example.com/miniapp" {
		t.Fatalf("expected button URL to be set")
	}
}

func TestMiniAppKeyboardWithoutURL(t *testing.T) {
	keyboard := MiniAppKeyboard("")
	if len(keyboard.InlineKeyboard) != 0 {
		t.Fatalf("expected no keyboard rows, got %d", len(keyboard.InlineKeyboard))
	}
}
