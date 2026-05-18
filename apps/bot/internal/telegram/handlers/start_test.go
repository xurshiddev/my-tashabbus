package handlers

import (
	"io"
	"log/slog"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestStartHandlerUsesChairmanMessageForOwner(t *testing.T) {
	sender := &captureSender{}
	handler := NewStartHandler(sender, "https://miniapp.example", 123, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	handler.Handle(startUpdate(123))

	if !strings.Contains(sender.text, "MFY raisi") {
		t.Fatalf("expected chairman message, got %q", sender.text)
	}
}

func TestStartHandlerIncludesMyIDInstructionForUnassignedUser(t *testing.T) {
	sender := &captureSender{}
	handler := NewStartHandler(sender, "https://miniapp.example", 123, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	handler.Handle(startUpdate(456))

	if !strings.Contains(sender.text, "/myid") {
		t.Fatalf("expected /myid instruction, got %q", sender.text)
	}
}

type captureSender struct {
	text string
}

func (s *captureSender) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	messageConfig, ok := c.(tgbotapi.MessageConfig)
	if ok {
		s.text = messageConfig.Text
	}
	return tgbotapi.Message{}, nil
}

func startUpdate(userID int64) tgbotapi.Update {
	return tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 1},
			From: &tgbotapi.User{ID: userID},
		},
	}
}
