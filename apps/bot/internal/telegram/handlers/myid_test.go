package handlers

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestFormatMyIDMessageWithUsername(t *testing.T) {
	message := FormatMyIDMessage(&tgbotapi.User{ID: 123456789, UserName: "tester"})

	expected := "Sizning Telegram ID: 123456789\nUsername: tester"
	if message != expected {
		t.Fatalf("expected %q, got %q", expected, message)
	}
}

func TestFormatMyIDMessageWithoutUsername(t *testing.T) {
	message := FormatMyIDMessage(&tgbotapi.User{ID: 123456789})

	expected := "Sizning Telegram ID: 123456789\nUsername: "
	if message != expected {
		t.Fatalf("expected %q, got %q", expected, message)
	}
}
