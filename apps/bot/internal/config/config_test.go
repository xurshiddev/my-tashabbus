package config

import "testing"

func TestValidateRequiresBotToken(t *testing.T) {
	cfg := Config{}
	if err := cfg.Validate(); err != ErrMissingBotToken {
		t.Fatalf("expected ErrMissingBotToken, got %v", err)
	}
}
