package config

import "testing"

func TestMFYChairmanTelegramIDEnvResolution(t *testing.T) {
	t.Setenv("MFY_CHAIRMAN_TELEGRAM_ID", "111")
	t.Setenv("ADMIN_TELEGRAM_ID", "222")
	t.Setenv("MFY_OWNER_TELEGRAM_ID", "333")

	cfg := Load()

	if cfg.MFYOwnerTelegramID != 111 {
		t.Fatalf("expected preferred chairman id 111, got %d", cfg.MFYOwnerTelegramID)
	}
}

func TestAdminTelegramIDFallback(t *testing.T) {
	t.Setenv("ADMIN_TELEGRAM_ID", "222")
	t.Setenv("MFY_OWNER_TELEGRAM_ID", "333")

	cfg := Load()

	if cfg.MFYOwnerTelegramID != 222 {
		t.Fatalf("expected admin id fallback 222, got %d", cfg.MFYOwnerTelegramID)
	}
}

func TestMFYOwnerTelegramIDFallback(t *testing.T) {
	t.Setenv("MFY_OWNER_TELEGRAM_ID", "333")

	cfg := Load()

	if cfg.MFYOwnerTelegramID != 333 {
		t.Fatalf("expected legacy owner id fallback 333, got %d", cfg.MFYOwnerTelegramID)
	}
}
